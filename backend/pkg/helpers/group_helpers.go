package helpers

import (
	"errors"
	"fmt"
	"time"

	"github.com/HASANALI117/social-network/pkg/db"
	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/google/uuid"
)

var (
	ErrGroupNotFound      = errors.New("group not found")
	ErrGroupAlreadyExists = errors.New("group with this name already exists")
	ErrNotGroupMember     = errors.New("user is not a member of this group")
	ErrAlreadyGroupMember = errors.New("user is already a member of this group")
)

// CreateGroup creates a new group
func CreateGroup(group *models.Group) error {
	query := `
        INSERT INTO groups (id, name, description, creator_id, avatar_url, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `

	group.ID = uuid.New().String()
	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()

	_, err := db.GlobalDB.Exec(
		query,
		group.ID,
		group.Name,
		group.Description,
		group.CreatorID,
		group.AvatarURL,
		group.CreatedAt,
		group.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	// Add creator as admin
	memberQuery := `
        INSERT INTO group_members (group_id, user_id, role, joined_at)
        VALUES (?, ?, 'admin', ?)
    `
	_, err = db.GlobalDB.Exec(memberQuery, group.ID, group.CreatorID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to add creator as admin: %w", err)
	}

	return nil
}

// GetGroupByID retrieves a group by ID
func GetGroupByID(id string) (*models.Group, error) {
	query := `
        SELECT id, name, description, creator_id, avatar_url, created_at, updated_at
        FROM groups
        WHERE id = ?
    `

	group := &models.Group{}
	err := db.GlobalDB.QueryRow(query, id).Scan(
		&group.ID,
		&group.Name,
		&group.Description,
		&group.CreatorID,
		&group.AvatarURL,
		&group.CreatedAt,
		&group.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	return group, nil
}

// UpdateGroup updates an existing group
func UpdateGroup(group *models.Group) error {
	group.UpdatedAt = time.Now()

	query := `
        UPDATE groups
        SET name = ?, description = ?, avatar_url = ?, updated_at = ?
        WHERE id = ?
    `

	result, err := db.GlobalDB.Exec(
		query,
		group.Name,
		group.Description,
		group.AvatarURL,
		group.UpdatedAt,
		group.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrGroupNotFound
	}

	return nil
}

// DeleteGroup deletes a group
func DeleteGroup(id string) error {
	query := "DELETE FROM groups WHERE id = ?"
	result, err := db.GlobalDB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrGroupNotFound
	}

	return nil
}

// ListGroups returns a list of all groups with pagination
func ListGroups(limit, offset int) ([]*models.Group, error) {
	query := `
        SELECT id, name, description, creator_id, avatar_url, created_at, updated_at
        FROM groups
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `

	rows, err := db.GlobalDB.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}
	defer rows.Close()

	groups := make([]*models.Group, 0)
	for rows.Next() {
		group := &models.Group{}
		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.Description,
			&group.CreatorID,
			&group.AvatarURL,
			&group.CreatedAt,
			&group.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}

	return groups, nil
}

// ListGroupsByUser returns groups a user is a member of
func ListGroupsByUser(userID string, limit, offset int) ([]*models.Group, error) {
	query := `
        SELECT g.id, g.name, g.description, g.creator_id, g.avatar_url, g.created_at, g.updated_at
        FROM groups g
        JOIN group_members gm ON g.id = gm.group_id
        WHERE gm.user_id = ?
        ORDER BY g.created_at DESC
        LIMIT ? OFFSET ?
    `

	rows, err := db.GlobalDB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user's groups: %w", err)
	}
	defer rows.Close()

	groups := make([]*models.Group, 0)
	for rows.Next() {
		group := &models.Group{}
		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.Description,
			&group.CreatorID,
			&group.AvatarURL,
			&group.CreatedAt,
			&group.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}

	return groups, nil
}

// AddGroupMember adds a user to a group
func AddGroupMember(groupID, userID, role string) error {
	// Check if user is already a member
	var exists bool
	checkQuery := "SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ? LIMIT 1"
	err := db.GlobalDB.QueryRow(checkQuery, groupID, userID).Scan(&exists)
	if err == nil {
		return ErrAlreadyGroupMember
	}

	// Add member
	query := `
        INSERT INTO group_members (group_id, user_id, role, joined_at)
        VALUES (?, ?, ?, ?)
    `
	if role == "" {
		role = "member"
	}

	_, err = db.GlobalDB.Exec(query, groupID, userID, role, time.Now())
	if err != nil {
		return fmt.Errorf("failed to add group member: %w", err)
	}

	return nil
}

// RemoveGroupMember removes a user from a group
func RemoveGroupMember(groupID, userID string) error {
	query := "DELETE FROM group_members WHERE group_id = ? AND user_id = ?"
	result, err := db.GlobalDB.Exec(query, groupID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove group member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotGroupMember
	}

	return nil
}

// ListGroupMembers returns members of a group
func ListGroupMembers(groupID string) ([]*models.User, error) {
	query := `
        SELECT u.id, u.username, u.email, u.password_hash, u.first_name, u.last_name, 
               u.avatar_url, u.about_me, u.birth_date, u.created_at, u.updated_at
        FROM users u
        JOIN group_members gm ON u.id = gm.user_id
        WHERE gm.group_id = ?
        ORDER BY gm.joined_at
    `

	rows, err := db.GlobalDB.Query(query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to list group members: %w", err)
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.FirstName,
			&user.LastName,
			&user.AvatarURL,
			&user.AboutMe,
			&user.BirthDate,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// SaveGroupMessage saves a message to a group
func SaveGroupMessage(message *models.GroupMessage) error {
	query := `
        INSERT INTO group_messages (id, group_id, sender_id, content, created_at)
        VALUES (?, ?, ?, ?, ?)
    `

	message.ID = uuid.NewString()
	message.CreatedAt = time.Now() // Use time.Time directly

	_, err := db.GlobalDB.Exec(
		query,
		message.ID,
		message.GroupID,
		message.SenderID,
		message.Content,
		message.CreatedAt,
	)

	return err
}

// GetGroupMessages retrieves messages for a group
func GetGroupMessages(groupID string, limit, offset int) ([]*models.GroupMessage, error) {
	query := `
        SELECT id, group_id, sender_id, content, created_at
        FROM group_messages
        WHERE group_id = ?
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `

	rows, err := db.GlobalDB.Query(query, groupID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]*models.GroupMessage, 0)
	for rows.Next() {
		msg := &models.GroupMessage{}
		err := rows.Scan(
			&msg.ID,
			&msg.GroupID,
			&msg.SenderID,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// IsGroupMember checks if a user is a member of a group
func IsGroupMember(groupID, userID string) (bool, error) {
	query := "SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ? LIMIT 1"
	var exists int
	err := db.GlobalDB.QueryRow(query, groupID, userID).Scan(&exists)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// IsGroupAdmin checks if a user is an admin of a group
func IsGroupAdmin(groupID, userID string) (bool, error) {
	query := "SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ? AND role = 'admin' LIMIT 1"
	var exists int
	err := db.GlobalDB.QueryRow(query, groupID, userID).Scan(&exists)
	if err != nil {
		return false, nil
	}
	return true, nil
}
