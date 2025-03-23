package helpers

import (
	"time"

	"social-network/pkg/db"
	"social-network/pkg/models"
	"github.com/google/uuid"
)

func CreateNotification(userID, typ, relatedID, message string) error {
	query := `
        INSERT INTO notifications (id, user_id, type, related_id, message, created_at)
        VALUES (?, ?, ?, ?, ?, ?)
    `
	_, err := db.GlobalDB.Exec(query, uuid.New().String(), userID, typ, relatedID, message, time.Now())
	return err
}

func ListNotifications(userID string) ([]*models.Notification, error) {
	query := `
        SELECT id, user_id, type, related_id, message, is_read, created_at
        FROM notifications
        WHERE user_id = ?
        ORDER BY created_at DESC
    `
	rows, err := db.GlobalDB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		n := &models.Notification{}
		err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.RelatedID, &n.Message, &n.IsRead, &n.CreatedAt)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}
