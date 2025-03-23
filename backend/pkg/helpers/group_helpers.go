package helpers

import (
	"time"

	"social-network/pkg/db"
	"social-network/pkg/models"
)

func CreateGroup(group *models.Group) error {
	query := `
        INSERT INTO groups (id, title, description, creator_id, created_at)
        VALUES (?, ?, ?, ?, ?)
    `
	_, err := db.GlobalDB.Exec(query, group.ID, group.Title, group.Description, group.CreatorID, group.CreatedAt)
	return err
}

func AddGroupMember(groupID, userID, status string) error {
	query := `
        INSERT INTO group_members (group_id, user_id, status, created_at)
        VALUES (?, ?, ?, ?)
    `
	_, err := db.GlobalDB.Exec(query, groupID, userID, status, time.Now())
	return err
}
