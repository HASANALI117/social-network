package repositories

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/google/uuid"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *models.Notification) error
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error)
	MarkAsRead(ctx context.Context, notificationID string, userID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}

type notificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	notification.ID = uuid.NewString()
	notification.CreatedAt = time.Now()
	notification.IsRead = false

	query := `INSERT INTO notifications (id, user_id, type, entity_type, message, entity_id, is_read, created_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query, notification.ID, notification.UserID, notification.Type, notification.EntityType, notification.Message, notification.EntityID, notification.IsRead, notification.CreatedAt)
	if err != nil {
		log.Printf("Error creating notification: %v", err)
		return err
	}
	return nil
}

func (r *notificationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error) {
	query := `SELECT id, user_id, type, entity_type, message, entity_id, is_read, created_at
              FROM notifications
              WHERE user_id = $1
              ORDER BY created_at DESC
              LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		log.Printf("Error getting notifications by user ID %s: %v", userID, err)
		return nil, err
	}
	defer rows.Close()

	notifications := make([]*models.Notification, 0)
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.EntityType, &n.Message, &n.EntityID, &n.IsRead, &n.CreatedAt); err != nil {
			log.Printf("Error scanning notification row: %v", err)
			return nil, err
		}
		notifications = append(notifications, &n)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating notification rows: %v", err)
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, notificationID string, userID string) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, notificationID, userID)
	if err != nil {
		log.Printf("Error marking notification %s as read for user %s: %v", notificationID, userID, err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected for MarkAsRead: %v", err)
		return err
	}
	if rowsAffected == 0 {
		log.Printf("No notification found with ID %s for user %s to mark as read", notificationID, userID)
		return sql.ErrNoRows // Or a custom error
	}
	return nil
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("Error marking all notifications as read for user %s: %v", userID, err)
		return err
	}
	return nil
}

func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		log.Printf("Error getting unread notification count for user %s: %v", userID, err)
		return 0, err
	}
	return count, nil
}