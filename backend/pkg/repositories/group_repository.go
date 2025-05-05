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
	// ErrInvitationNotFound indicates that a group invitation was not found.
	ErrInvitationNotFound = errors.New("group invitation not found")
	// ErrJoinRequestNotFound indicates that a group join request was not found.
	ErrJoinRequestNotFound = errors.New("group join request not found")
	// ErrAlreadyInvited indicates a user has already been invited to the group.
	ErrAlreadyInvited = errors.New("user has already been invited to this group")
	// ErrAlreadyRequested indicates a user has already requested to join the group.
	ErrAlreadyRequested = errors.New("user has already requested to join this group")
)

// GroupRepository defines the interface for group data access
type GroupRepository interface {
	// Group CRUD
	Create(group *models.Group) error
	GetByID(id string) (*models.Group, error)
	List(limit, offset int, searchQuery string) ([]*models.Group, error) // Added searchQuery
	Update(group *models.Group) error
	Delete(id string) error

	// Member Management
	AddMember(groupID, userID, role string) error
	RemoveMember(groupID, userID string) error
	ListMembers(groupID string) ([]*models.User, error) // Returns User models
	IsMember(groupID, userID string) (bool, error)
	IsAdmin(groupID, userID string) (bool, error)
	ListGroupsByUser(userID string, limit, offset int) ([]*models.Group, error) // Added

	// Invitation Management
	CreateInvitation(invitation *models.GroupInvitation) error
	GetInvitationByID(invitationID string) (*models.GroupInvitation, error)
	FindPendingInvitation(groupID, inviteeID string) (*models.GroupInvitation, error) // Find specific pending invite
	UpdateInvitationStatus(invitationID, status string) error
	ListPendingInvitationsForUser(inviteeID string) ([]*models.GroupInvitation, error) // List invites received by user
	ListPendingInvitationsForGroup(groupID string) ([]*models.GroupInvitation, error)  // List invites sent by group members
	DeleteInvitation(invitationID string) error

	// Join Request Management
	CreateJoinRequest(request *models.GroupJoinRequest) error
	GetJoinRequestByID(requestID string) (*models.GroupJoinRequest, error)
	FindPendingJoinRequest(groupID, requesterID string) (*models.GroupJoinRequest, error) // Find specific pending request
	UpdateJoinRequestStatus(requestID, status string) error
	ListPendingJoinRequestsForGroup(groupID string) ([]*models.GroupJoinRequest, error) // List requests for a group (for creator/admins)
	DeleteJoinRequest(requestID string) error

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

// List retrieves a paginated list of all groups, optionally filtered by search query
func (r *groupRepository) List(limit, offset int, searchQuery string) ([]*models.Group, error) {
	baseQuery := `
        SELECT id, creator_id, name, description, avatar_url, created_at, updated_at
        FROM groups
    `
	args := []interface{}{}
	whereClauses := []string{}

	if searchQuery != "" {
		// Add WHERE clause for search (case-insensitive partial match)
		// Use LOWER() for case-insensitivity, works in SQLite and PostgreSQL
		whereClauses = append(whereClauses, "(LOWER(name) LIKE ? OR LOWER(description) LIKE ?)")
		searchTerm := "%" + strings.ToLower(searchQuery) + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Construct the final query
	query := baseQuery
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ") // Use AND if more clauses are added later
	}
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups with query '%s': %w", query, err)
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

// ListGroupsByUser returns groups a user is a member of with pagination
func (r *groupRepository) ListGroupsByUser(userID string, limit, offset int) ([]*models.Group, error) {
	query := `
	       SELECT g.id, g.name, g.description, g.creator_id, g.avatar_url, g.created_at, g.updated_at
	       FROM groups g
	       JOIN group_members gm ON g.id = gm.group_id
	       WHERE gm.user_id = ?
	       ORDER BY g.created_at DESC
	       LIMIT ? OFFSET ?
	   `

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user's groups: %w", err)
	}
	defer rows.Close()

	groups := make([]*models.Group, 0)
	for rows.Next() {
		group := &models.Group{}
		var createdAt, updatedAt sql.NullString // Use NullString for nullable timestamps
		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.Description,
			&group.CreatorID,
			&group.AvatarURL,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group during ListGroupsByUser: %w", err)
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
		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user's group list rows: %w", err)
	}

	return groups, nil
}

// --- Invitation Management ---

// CreateInvitation inserts a new group invitation record.
func (r *groupRepository) CreateInvitation(invitation *models.GroupInvitation) error {
	query := `
        INSERT INTO group_invitations (id, group_id, inviter_id, invitee_id, status, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `
	invitation.ID = uuid.New().String()
	invitation.Status = "pending" // Ensure status is pending on creation
	now := time.Now()
	invitation.CreatedAt = now
	invitation.UpdatedAt = now

	_, err := r.db.Exec(
		query,
		invitation.ID,
		invitation.GroupID,
		invitation.InviterID,
		invitation.InviteeID,
		invitation.Status,
		invitation.CreatedAt,
		invitation.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: group_invitations.group_id, group_invitations.invitee_id") {
			return ErrAlreadyInvited
		}
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			// Could be group, inviter, or invitee not found
			return fmt.Errorf("failed to create invitation: invalid group, inviter, or invitee ID: %w", err)
		}
		return fmt.Errorf("failed to create group invitation: %w", err)
	}
	return nil
}

// GetInvitationByID retrieves a group invitation by its ID.
func (r *groupRepository) GetInvitationByID(invitationID string) (*models.GroupInvitation, error) {
	query := `
        SELECT id, group_id, inviter_id, invitee_id, status, created_at, updated_at
        FROM group_invitations
        WHERE id = ?
    `
	var inv models.GroupInvitation
	var createdAtStr, updatedAtStr string // Scan into strings first

	err := r.db.QueryRow(query, invitationID).Scan(
		&inv.ID,
		&inv.GroupID,
		&inv.InviterID,
		&inv.InviteeID,
		&inv.Status,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvitationNotFound
		}
		return nil, fmt.Errorf("failed to get group invitation by ID: %w", err)
	}

	// Parse timestamps
	inv.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		fmt.Printf("Warning: Failed to parse invitation created_at timestamp '%s': %v\n", createdAtStr, err)
	}
	inv.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		fmt.Printf("Warning: Failed to parse invitation updated_at timestamp '%s': %v\n", updatedAtStr, err)
	}

	return &inv, nil
}

// FindPendingInvitation retrieves a specific pending invitation for a user to a group.
func (r *groupRepository) FindPendingInvitation(groupID, inviteeID string) (*models.GroupInvitation, error) {
	query := `
        SELECT id, group_id, inviter_id, invitee_id, status, created_at, updated_at
        FROM group_invitations
        WHERE group_id = ? AND invitee_id = ? AND status = 'pending'
    `
	var inv models.GroupInvitation
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRow(query, groupID, inviteeID).Scan(
		&inv.ID,
		&inv.GroupID,
		&inv.InviterID,
		&inv.InviteeID,
		&inv.Status,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvitationNotFound // Or maybe just nil, nil?
		}
		return nil, fmt.Errorf("failed to find pending group invitation: %w", err)
	}

	// Parse timestamps
	inv.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		fmt.Printf("Warning: Failed to parse invitation created_at timestamp '%s': %v\n", createdAtStr, err)
	}
	inv.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		fmt.Printf("Warning: Failed to parse invitation updated_at timestamp '%s': %v\n", updatedAtStr, err)
	}

	return &inv, nil
}

// UpdateInvitationStatus updates the status of a group invitation.
func (r *groupRepository) UpdateInvitationStatus(invitationID, status string) error {
	query := `
        UPDATE group_invitations
        SET status = ?, updated_at = CURRENT_TIMESTAMP
        WHERE id = ?
    `
	result, err := r.db.Exec(query, status, invitationID)
	if err != nil {
		return fmt.Errorf("failed to update group invitation status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after updating invitation status: %w", err)
	}
	if rowsAffected == 0 {
		return ErrInvitationNotFound
	}
	return nil
}

// ListPendingInvitationsForUser retrieves all pending invitations for a specific user.
func (r *groupRepository) ListPendingInvitationsForUser(inviteeID string) ([]*models.GroupInvitation, error) {
	query := `
        SELECT id, group_id, inviter_id, invitee_id, status, created_at, updated_at
        FROM group_invitations
        WHERE invitee_id = ? AND status = 'pending'
        ORDER BY created_at DESC
    `
	rows, err := r.db.Query(query, inviteeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending invitations for user: %w", err)
	}
	defer rows.Close()

	invitations := make([]*models.GroupInvitation, 0)
	for rows.Next() {
		var inv models.GroupInvitation
		var createdAtStr, updatedAtStr string
		err := rows.Scan(
			&inv.ID,
			&inv.GroupID,
			&inv.InviterID,
			&inv.InviteeID,
			&inv.Status,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation during list: %w", err)
		}
		// Parse timestamps
		inv.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			fmt.Printf("Warning: Failed to parse invitation created_at timestamp '%s': %v\n", createdAtStr, err)
		}
		inv.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			fmt.Printf("Warning: Failed to parse invitation updated_at timestamp '%s': %v\n", updatedAtStr, err)
		}
		invitations = append(invitations, &inv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pending invitation list rows: %w", err)
	}

	return invitations, nil
}

// ListPendingInvitationsForGroup retrieves all pending invitations sent for a specific group.
func (r *groupRepository) ListPendingInvitationsForGroup(groupID string) ([]*models.GroupInvitation, error) {
	query := `
        SELECT id, group_id, inviter_id, invitee_id, status, created_at, updated_at
        FROM group_invitations
        WHERE group_id = ? AND status = 'pending'
        ORDER BY created_at DESC
    `
	rows, err := r.db.Query(query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending invitations for group: %w", err)
	}
	defer rows.Close()

	invitations := make([]*models.GroupInvitation, 0)
	for rows.Next() {
		var inv models.GroupInvitation
		var createdAtStr, updatedAtStr string
		err := rows.Scan(
			&inv.ID,
			&inv.GroupID,
			&inv.InviterID,
			&inv.InviteeID,
			&inv.Status,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation during list for group: %w", err)
		}
		// Parse timestamps
		inv.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			fmt.Printf("Warning: Failed to parse invitation created_at timestamp '%s': %v\n", createdAtStr, err)
		}
		inv.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			fmt.Printf("Warning: Failed to parse invitation updated_at timestamp '%s': %v\n", updatedAtStr, err)
		}
		invitations = append(invitations, &inv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pending invitation list rows for group: %w", err)
	}

	return invitations, nil
}

// DeleteInvitation removes a group invitation record.
func (r *groupRepository) DeleteInvitation(invitationID string) error {
	query := "DELETE FROM group_invitations WHERE id = ?"
	result, err := r.db.Exec(query, invitationID)
	if err != nil {
		return fmt.Errorf("failed to delete group invitation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after deleting invitation: %w", err)
	}
	if rowsAffected == 0 {
		return ErrInvitationNotFound
	}
	return nil
}

// --- Join Request Management ---

// CreateJoinRequest inserts a new group join request record.
func (r *groupRepository) CreateJoinRequest(request *models.GroupJoinRequest) error {
	query := `
        INSERT INTO group_join_requests (id, group_id, requester_id, status, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?)
    `
	request.ID = uuid.New().String()
	request.Status = "pending" // Ensure status is pending on creation
	now := time.Now()
	request.CreatedAt = now
	request.UpdatedAt = now

	_, err := r.db.Exec(
		query,
		request.ID,
		request.GroupID,
		request.RequesterID,
		request.Status,
		request.CreatedAt,
		request.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: group_join_requests.group_id, group_join_requests.requester_id") {
			return ErrAlreadyRequested
		}
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			// Could be group or requester not found
			return fmt.Errorf("failed to create join request: invalid group or requester ID: %w", err)
		}
		return fmt.Errorf("failed to create group join request: %w", err)
	}
	return nil
}

// GetJoinRequestByID retrieves a group join request by its ID.
func (r *groupRepository) GetJoinRequestByID(requestID string) (*models.GroupJoinRequest, error) {
	query := `
        SELECT id, group_id, requester_id, status, created_at, updated_at
        FROM group_join_requests
        WHERE id = ?
    `
	var req models.GroupJoinRequest
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRow(query, requestID).Scan(
		&req.ID,
		&req.GroupID,
		&req.RequesterID,
		&req.Status,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrJoinRequestNotFound
		}
		return nil, fmt.Errorf("failed to get group join request by ID: %w", err)
	}

	// Parse timestamps
	req.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		fmt.Printf("Warning: Failed to parse join request created_at timestamp '%s': %v\n", createdAtStr, err)
	}
	req.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		fmt.Printf("Warning: Failed to parse join request updated_at timestamp '%s': %v\n", updatedAtStr, err)
	}

	return &req, nil
}

// FindPendingJoinRequest retrieves a specific pending join request for a user to a group.
func (r *groupRepository) FindPendingJoinRequest(groupID, requesterID string) (*models.GroupJoinRequest, error) {
	query := `
        SELECT id, group_id, requester_id, status, created_at, updated_at
        FROM group_join_requests
        WHERE group_id = ? AND requester_id = ? AND status = 'pending'
    `
	var req models.GroupJoinRequest
	var createdAtStr, updatedAtStr string

	err := r.db.QueryRow(query, groupID, requesterID).Scan(
		&req.ID,
		&req.GroupID,
		&req.RequesterID,
		&req.Status,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrJoinRequestNotFound // Or maybe just nil, nil?
		}
		return nil, fmt.Errorf("failed to find pending group join request: %w", err)
	}

	// Parse timestamps
	req.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		fmt.Printf("Warning: Failed to parse join request created_at timestamp '%s': %v\n", createdAtStr, err)
	}
	req.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		fmt.Printf("Warning: Failed to parse join request updated_at timestamp '%s': %v\n", updatedAtStr, err)
	}

	return &req, nil
}

// UpdateJoinRequestStatus updates the status of a group join request.
func (r *groupRepository) UpdateJoinRequestStatus(requestID, status string) error {
	query := `
        UPDATE group_join_requests
        SET status = ?, updated_at = CURRENT_TIMESTAMP
        WHERE id = ?
    `
	result, err := r.db.Exec(query, status, requestID)
	if err != nil {
		return fmt.Errorf("failed to update group join request status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after updating join request status: %w", err)
	}
	if rowsAffected == 0 {
		return ErrJoinRequestNotFound
	}
	return nil
}

// ListPendingJoinRequestsForGroup retrieves all pending join requests for a specific group.
func (r *groupRepository) ListPendingJoinRequestsForGroup(groupID string) ([]*models.GroupJoinRequest, error) {
	query := `
        SELECT id, group_id, requester_id, status, created_at, updated_at
        FROM group_join_requests
        WHERE group_id = ? AND status = 'pending'
        ORDER BY created_at DESC
    `
	rows, err := r.db.Query(query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending join requests for group: %w", err)
	}
	defer rows.Close()

	requests := make([]*models.GroupJoinRequest, 0)
	for rows.Next() {
		var req models.GroupJoinRequest
		var createdAtStr, updatedAtStr string
		err := rows.Scan(
			&req.ID,
			&req.GroupID,
			&req.RequesterID,
			&req.Status,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan join request during list for group: %w", err)
		}
		// Parse timestamps
		req.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			fmt.Printf("Warning: Failed to parse join request created_at timestamp '%s': %v\n", createdAtStr, err)
		}
		req.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			fmt.Printf("Warning: Failed to parse join request updated_at timestamp '%s': %v\n", updatedAtStr, err)
		}
		requests = append(requests, &req)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pending join request list rows for group: %w", err)
	}

	return requests, nil
}

// DeleteJoinRequest removes a group join request record.
func (r *groupRepository) DeleteJoinRequest(requestID string) error {
	query := "DELETE FROM group_join_requests WHERE id = ?"
	result, err := r.db.Exec(query, requestID)
	if err != nil {
		return fmt.Errorf("failed to delete group join request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after deleting join request: %w", err)
	}
	if rowsAffected == 0 {
		return ErrJoinRequestNotFound
	}
	return nil
}
