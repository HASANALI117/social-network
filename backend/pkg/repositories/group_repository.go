package repositories

import (
"database/sql"
"errors"
"fmt"
"strings" // Import strings
"time"

"github.com/HASANALI117/social-network/pkg/models"
"github.com/google/uuid"
)

var (
// ErrGroupNotFound indicates that a group with the given ID was not found.
ErrGroupNotFound = errors.New("group not found")
// ErrAlreadyGroupMember indicates that a user is already a member of the group.
ErrAlreadyGroupMember = errors.New("user is already a member of this group")
// ErrNotGroupMember indicates that a user is not a member of the group.
ErrNotGroupMember = errors.New("user is not a member of this group")
)

// GroupRepository defines the interface for group data access
type GroupRepository interface {
Create(group *models.Group) error
GetByID(id string) (*models.Group, error)
List(limit, offset int) ([]*models.Group, error)
Update(group *models.Group) error
Delete(id string) error
AddMember(groupID, userID, role string) error
RemoveMember(groupID, userID string) error
ListMembers(groupID string) ([]*models.User, error) // Returns User models
IsMember(groupID, userID string) (bool, error)
IsAdmin(groupID, userID string) (bool, error)
// TODO: Add methods for group messages if needed here, or in a separate repo
}

// groupRepository implements GroupRepository interface
type groupRepository struct {
db *sql.DB
}

// NewGroupRepository creates a new GroupRepository
func NewGroupRepository(db *sql.DB) GroupRepository {
return &groupRepository{
db: db,
}
}

// Create inserts a new group record into the database
func (r *groupRepository) Create(group *models.Group) error {
tx, err := r.db.Begin()
if err != nil {
return fmt.Errorf("failed to begin transaction: %w", err)
}
defer tx.Rollback() // Rollback if anything fails

// Insert group
queryGroup := `
        INSERT INTO groups (id, creator_id, name, description, avatar_url, created_at)
        VALUES (?, ?, ?, ?, ?, ?)
    `
group.ID = uuid.New().String()
group.CreatedAt = time.Now()

_, err = tx.Exec(
queryGroup,
group.ID,
group.CreatorID,
group.Name,
group.Description,
group.AvatarURL,
group.CreatedAt,
)
if err != nil {
return fmt.Errorf("failed to create group: %w", err)
}

// Add creator as the first member (admin)
queryMember := `
        INSERT INTO group_members (group_id, user_id, role, joined_at)
        VALUES (?, ?, ?, ?)
    `
_, err = tx.Exec(queryMember, group.ID, group.CreatorID, "admin", time.Now())
if err != nil {
// Check for unique constraint violation (shouldn't happen for creator normally)
// but handle just in case
// Use strings.Contains for error checking
if strings.Contains(err.Error(), "UNIQUE constraint failed") {
return fmt.Errorf("failed to add creator as member (already exists?): %w", ErrAlreadyGroupMember)
}
return fmt.Errorf("failed to add creator as group member: %w", err)
}

if err := tx.Commit(); err != nil {
return fmt.Errorf("failed to commit transaction: %w", err)
}

return nil
}

// GetByID retrieves a group by its ID
func (r *groupRepository) GetByID(id string) (*models.Group, error) {
query := `
        SELECT id, creator_id, name, description, avatar_url, created_at, updated_at
        FROM groups
        WHERE id = ?
    `
var group models.Group
var createdAt, updatedAt sql.NullString // Use NullString for nullable updated_at

err := r.db.QueryRow(query, id).Scan(
&group.ID,
&group.CreatorID,
&group.Name,
&group.Description,
&group.AvatarURL,
&createdAt,
&updatedAt,
)
if err != nil {
if errors.Is(err, sql.ErrNoRows) {
return nil, ErrGroupNotFound
}
return nil, fmt.Errorf("failed to get group by ID: %w", err)
}

// Parse timestamps
if createdAt.Valid {
group.CreatedAt, err = time.Parse(time.RFC3339, createdAt.String)
if err != nil {
fmt.Printf("Warning: Failed to parse group created_at timestamp '%s': %v\n", createdAt.String, err)
group.CreatedAt = time.Time{}
}
}
if updatedAt.Valid {
group.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt.String)
if err != nil {
fmt.Printf("Warning: Failed to parse group updated_at timestamp '%s': %v\n", updatedAt.String, err)
group.UpdatedAt = time.Time{} // Or keep nil/zero?
}
}

return &group, nil
}

// List retrieves a paginated list of all groups
func (r *groupRepository) List(limit, offset int) ([]*models.Group, error) {
query := `
        SELECT id, creator_id, name, description, avatar_url, created_at, updated_at
        FROM groups
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `
rows, err := r.db.Query(query, limit, offset)
if err != nil {
return nil, fmt.Errorf("failed to list groups: %w", err)
}
defer rows.Close()

groups := make([]*models.Group, 0)
for rows.Next() {
var group models.Group
var createdAt, updatedAt sql.NullString
err := rows.Scan(
&group.ID,
&group.CreatorID,
&group.Name,
&group.Description,
&group.AvatarURL,
&createdAt,
&updatedAt,
)
if err != nil {
return nil, fmt.Errorf("failed to scan group during list: %w", err)
}
// Parse timestamps
if createdAt.Valid {
group.CreatedAt, err = time.Parse(time.RFC3339, createdAt.String)
if err != nil {
fmt.Printf("Warning: Failed to parse group created_at timestamp '%s': %v\n", createdAt.String, err)
group.CreatedAt = time.Time{}
}
}
if updatedAt.Valid {
group.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt.String)
if err != nil {
fmt.Printf("Warning: Failed to parse group updated_at timestamp '%s': %v\n", updatedAt.String, err)
}
}
groups = append(groups, &group)
}

if err := rows.Err(); err != nil {
return nil, fmt.Errorf("error iterating group list rows: %w", err)
}

return groups, nil
}

// Update modifies an existing group record
func (r *groupRepository) Update(group *models.Group) error {
query := `
        UPDATE groups
        SET name = ?, description = ?, avatar_url = ?, updated_at = ?
        WHERE id = ?
    `
group.UpdatedAt = time.Now()
result, err := r.db.Exec(
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
return fmt.Errorf("failed to get rows affected after updating group: %w", err)
}

if rowsAffected == 0 {
return ErrGroupNotFound // Return error if no rows were updated
}

return nil
}

// Delete removes a group and its related data (members, messages - cascade or manual)
func (r *groupRepository) Delete(id string) error {
// Using CASCADE DELETE defined in schema is simpler.
// If not using CASCADE, delete members and messages manually within a transaction.
tx, err := r.db.Begin()
if err != nil {
return fmt.Errorf("failed to begin transaction for group delete: %w", err)
}
defer tx.Rollback()

// Delete group (assuming cascade delete handles members/messages)
query := "DELETE FROM groups WHERE id = ?"
result, err := tx.Exec(query, id)
if err != nil {
return fmt.Errorf("failed to delete group: %w", err)
}

rowsAffected, err := result.RowsAffected()
if err != nil {
return fmt.Errorf("failed to get rows affected after deleting group: %w", err)
}

if rowsAffected == 0 {
return ErrGroupNotFound // Return error if no rows were deleted
}

if err := tx.Commit(); err != nil {
return fmt.Errorf("failed to commit transaction for group delete: %w", err)
}

return nil
}

// AddMember adds a user to a group
func (r *groupRepository) AddMember(groupID, userID, role string) error {
if role == "" {
role = "member" // Default role
}
query := `
        INSERT INTO group_members (group_id, user_id, role, joined_at)
        VALUES (?, ?, ?, ?)
    `
_, err := r.db.Exec(query, groupID, userID, role, time.Now())
if err != nil {
// Check for unique constraint violation
// Use strings.Contains for error checking
if strings.Contains(err.Error(), "UNIQUE constraint failed") {
return ErrAlreadyGroupMember
}
// Check for foreign key violation (group or user doesn't exist)
// Use strings.Contains for error checking
if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
// Could check which FK failed, but returning a generic error might be okay
return fmt.Errorf("failed to add member: group or user not found")
}
return fmt.Errorf("failed to add group member: %w", err)
}
return nil
}

// RemoveMember removes a user from a group
func (r *groupRepository) RemoveMember(groupID, userID string) error {
query := "DELETE FROM group_members WHERE group_id = ? AND user_id = ?"
result, err := r.db.Exec(query, groupID, userID)
if err != nil {
return fmt.Errorf("failed to remove group member: %w", err)
}

rowsAffected, err := result.RowsAffected()
if err != nil {
return fmt.Errorf("failed to get rows affected after removing member: %w", err)
}

if rowsAffected == 0 {
return ErrNotGroupMember // Return error if no rows were deleted
}

return nil
}

// ListMembers retrieves all users who are members of a group
func (r *groupRepository) ListMembers(groupID string) ([]*models.User, error) {
query := `
        SELECT u.id, u.username, u.email, u.first_name, u.last_name, u.avatar_url, u.about_me, u.created_at
        FROM users u
        JOIN group_members gm ON u.id = gm.user_id
        WHERE gm.group_id = ?
        ORDER BY gm.joined_at
    `
rows, err := r.db.Query(query, groupID)
if err != nil {
return nil, fmt.Errorf("failed to list group members: %w", err)
}
defer rows.Close()

members := make([]*models.User, 0)
for rows.Next() {
var user models.User
var createdAt string
err := rows.Scan(
&user.ID,
&user.Username,
&user.Email,
&user.FirstName,
&user.LastName,
&user.AvatarURL,
&user.AboutMe,
&createdAt,
)
if err != nil {
return nil, fmt.Errorf("failed to scan member during list: %w", err)
}
// Parse timestamp
user.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
if err != nil {
fmt.Printf("Warning: Failed to parse member created_at timestamp '%s': %v\n", createdAt, err)
user.CreatedAt = time.Time{}
}
members = append(members, &user)
}

if err := rows.Err(); err != nil {
return nil, fmt.Errorf("error iterating member list rows: %w", err)
}

return members, nil
}

// IsMember checks if a user is a member of a group
func (r *groupRepository) IsMember(groupID, userID string) (bool, error) {
query := "SELECT EXISTS (SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ?)"
var exists bool
err := r.db.QueryRow(query, groupID, userID).Scan(&exists)
if err != nil {
return false, fmt.Errorf("failed to check group membership: %w", err)
}
return exists, nil
}

// IsAdmin checks if a user is an admin of a group
func (r *groupRepository) IsAdmin(groupID, userID string) (bool, error) {
query := "SELECT EXISTS (SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ? AND role = 'admin')"
var isAdmin bool
err := r.db.QueryRow(query, groupID, userID).Scan(&isAdmin)
if err != nil {
return false, fmt.Errorf("failed to check group admin status: %w", err)
}
return isAdmin, nil
}
