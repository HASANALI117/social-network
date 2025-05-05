package repositories

import (
	"database/sql"
	"log"
	"time"

	"github.com/HASANALI117/social-network/pkg/models"
)

// NotificationRepository defines the interface for notification data operations.
type NotificationRepository interface {
	Create(notification *models.Notification) error
	GetByUserID(userID int, limit, offset int) ([]models.Notification, error)
	MarkAsRead(notificationID int, userID int) error
	MarkAllAsRead(userID int) error
	GetUnreadCount(userID int) (int, error)
}

// sqliteNotificationRepository implements NotificationRepository for SQLite.
type sqliteNotificationRepository struct {
	db *sql.DB
}

// NewSQLiteNotificationRepository creates a new instance of sqliteNotificationRepository.
func NewSQLiteNotificationRepository(db *sql.DB) NotificationRepository {
	return &sqliteNotificationRepository{db: db}
}

// Create inserts a new notification record into the database.
func (r *sqliteNotificationRepository) Create(notification *models.Notification) error {
	query := `
        INSERT INTO notifications (user_id, actor_id, type, entity_id, entity_type, is_read, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?);
    `
	stmt, err := r.db.Prepare(query)
	if err != nil {
		log.Printf("Error preparing create notification statement: %v", err)
		return err
	}
	defer stmt.Close()

	// Use 0 for false (is_read)
	isRead := 0
	if notification.IsRead {
		isRead = 1
	}

	// Set CreatedAt if it's zero
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}

	// Handle nullable fields
	var actorID sql.NullInt64
	if notification.ActorID != nil {
		actorID = sql.NullInt64{Int64: int64(*notification.ActorID), Valid: true}
	}
	var entityID sql.NullInt64
	if notification.EntityID != nil {
		entityID = sql.NullInt64{Int64: int64(*notification.EntityID), Valid: true}
	}
	var entityType sql.NullString
	if notification.EntityType != nil {
		entityType = sql.NullString{String: string(*notification.EntityType), Valid: true}
	}

	res, err := stmt.Exec(
		notification.UserID,
		actorID,
		notification.Type,
		entityID,
		entityType,
		isRead,
		notification.CreatedAt,
	)
	if err != nil {
		log.Printf("Error executing create notification statement: %v", err)
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID for notification: %v", err)
		return err
	}
	notification.ID = int(id)

	return nil
}

// GetByUserID retrieves notifications for a specific user, ordered by creation date (newest first), with pagination.
func (r *sqliteNotificationRepository) GetByUserID(userID int, limit, offset int) ([]models.Notification, error) {
	query := `
        SELECT id, user_id, actor_id, type, entity_id, entity_type, is_read, created_at
        FROM notifications
        WHERE user_id = ?
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?;
    `
	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		log.Printf("Error querying notifications by user ID %d: %v", userID, err)
		return nil, err
	}
	defer rows.Close()

	notifications := []models.Notification{}
	for rows.Next() {
		var n models.Notification
		var isRead int // Use int for SQLite boolean
		var actorID sql.NullInt64
		var entityID sql.NullInt64
		var entityType sql.NullString

		err := rows.Scan(
			&n.ID,
			&n.UserID,
			&actorID,
			&n.Type,
			&entityID,
			&entityType,
			&isRead,
			&n.CreatedAt,
		)
		if err != nil {
			log.Printf("Error scanning notification row: %v", err)
			return nil, err
		}
		n.IsRead = (isRead == 1) // Convert int to bool

		// Assign values from nullable types if they are valid
		if actorID.Valid {
			actorIDVal := int(actorID.Int64)
			n.ActorID = &actorIDVal
		}
		if entityID.Valid {
			entityIDVal := int(entityID.Int64)
			n.EntityID = &entityIDVal
		}
		if entityType.Valid {
			entityTypeVal := models.EntityType(entityType.String)
			n.EntityType = &entityTypeVal
		}

		notifications = append(notifications, n)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating notification rows: %v", err)
		return nil, err
	}

	return notifications, nil
}

// MarkAsRead marks a specific notification as read for a given user.
func (r *sqliteNotificationRepository) MarkAsRead(notificationID int, userID int) error {
	query := `
        UPDATE notifications
        SET is_read = 1
        WHERE id = ? AND user_id = ? AND is_read = 0;
    `
	stmt, err := r.db.Prepare(query)
	if err != nil {
		log.Printf("Error preparing mark notification as read statement: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(notificationID, userID)
	if err != nil {
		log.Printf("Error executing mark notification %d as read for user %d: %v", notificationID, userID, err)
		return err
	}

	// Note: We don't check RowsAffected here. If the notification didn't exist,
	// wasn't owned by the user, or was already read, the query succeeds with 0 rows affected.
	// This is generally acceptable behavior for marking as read.

	return nil
}

// MarkAllAsRead marks all unread notifications for a specific user as read.
func (r *sqliteNotificationRepository) MarkAllAsRead(userID int) error {
	query := `
        UPDATE notifications
        SET is_read = 1
        WHERE user_id = ? AND is_read = 0;
    `
	stmt, err := r.db.Prepare(query)
	if err != nil {
		log.Printf("Error preparing mark all notifications as read statement: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID)
	if err != nil {
		log.Printf("Error executing mark all notifications as read for user %d: %v", userID, err)
		return err
	}

	return nil
}

// GetUnreadCount gets the count of unread notifications for a user.
func (r *sqliteNotificationRepository) GetUnreadCount(userID int) (int, error) {
	query := `
        SELECT COUNT(*)
        FROM notifications
        WHERE user_id = ? AND is_read = 0;
    `
	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			// This shouldn't happen with COUNT(*), but handle defensively
			return 0, nil
		}
		log.Printf("Error querying unread notification count for user %d: %v", userID, err)
		return 0, err
	}

	return count, nil
}
