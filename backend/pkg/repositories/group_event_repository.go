package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/HASANALI117/social-network/pkg/models"
)

var (
	// ErrEventNotFound indicates that an event with the given ID was not found
	ErrEventNotFound = errors.New("event not found")
	// ErrEventCreatorRequired indicates that only the creator can perform this operation
	ErrEventCreatorRequired = errors.New("only the event creator can perform this operation")
)

// GroupEventRepository defines the interface for group event data access
type GroupEventRepository interface {
	Create(event *models.GroupEvent) error
	GetByID(id string) (*models.GroupEvent, error)
	ListByGroupID(groupID string, limit, offset int) ([]*models.GroupEvent, error)
	Update(event *models.GroupEvent) error
	Delete(id string) error
}

// groupEventRepository implements GroupEventRepository interface
type groupEventRepository struct {
	db *sql.DB
}

// NewGroupEventRepository creates a new GroupEventRepository
func NewGroupEventRepository(db *sql.DB) GroupEventRepository {
	return &groupEventRepository{
		db: db,
	}
}

// Create inserts a new group event record into the database
func (r *groupEventRepository) Create(event *models.GroupEvent) error {
	query := `
        INSERT INTO group_events (id, group_id, creator_id, title, description, event_time, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `
	now := time.Now()
	event.CreatedAt = now
	event.UpdatedAt = now

	_, err := r.db.Exec(
		query,
		event.ID,
		event.GroupID,
		event.CreatorID,
		event.Title,
		event.Description,
		event.EventTime, // Pass time.Time directly
		event.CreatedAt, // Pass time.Time directly
		event.UpdatedAt, // Pass time.Time directly
	)
	if err != nil {
		if err.Error() == "FOREIGN KEY constraint failed" {
			return fmt.Errorf("failed to create event: group or creator doesn't exist: %w", err)
		}
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// GetByID retrieves a group event by its ID
func (r *groupEventRepository) GetByID(id string) (*models.GroupEvent, error) {
	query := `
        SELECT id, group_id, creator_id, title, description, event_time, created_at, updated_at
        FROM group_events
        WHERE id = ?
    `
	var event models.GroupEvent
	// Remove intermediate string variables for time

	err := r.db.QueryRow(query, id).Scan(
		&event.ID,
		&event.GroupID,
		&event.CreatorID,
		&event.Title,
		&event.Description,
		&event.EventTime, // Scan directly into time.Time field
		&event.CreatedAt, // Scan directly into time.Time field
		&event.UpdatedAt, // Scan directly into time.Time field
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get event by ID: %w", err)
	}

	// No need to parse timestamps manually anymore

	return &event, nil
}

// ListByGroupID retrieves a paginated list of events for a specific group
func (r *groupEventRepository) ListByGroupID(groupID string, limit, offset int) ([]*models.GroupEvent, error) {
	query := `
        SELECT id, group_id, creator_id, title, description, event_time, created_at, updated_at
        FROM group_events
        WHERE group_id = ?
        ORDER BY event_time ASC
        LIMIT ? OFFSET ?
    `
	rows, err := r.db.Query(query, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list events for group ID %s: %w", groupID, err)
	}
	defer rows.Close()

	events := make([]*models.GroupEvent, 0)
	for rows.Next() {
		var event models.GroupEvent
		// Remove intermediate string variables for time
		err := rows.Scan(
			&event.ID,
			&event.GroupID,
			&event.CreatorID,
			&event.Title,
			&event.Description,
			&event.EventTime, // Scan directly into time.Time field
			&event.CreatedAt, // Scan directly into time.Time field
			&event.UpdatedAt, // Scan directly into time.Time field
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event during list: %w", err)
		}

		// No need to parse timestamps manually anymore

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event list rows: %w", err)
	}

	return events, nil
}

// Update modifies an existing group event
func (r *groupEventRepository) Update(event *models.GroupEvent) error {
	query := `
        UPDATE group_events
        SET title = ?, description = ?, event_time = ?, updated_at = ?
        WHERE id = ?
    `
	event.UpdatedAt = time.Now()
	result, err := r.db.Exec(
		query,
		event.Title,
		event.Description,
		event.EventTime,
		event.UpdatedAt,
		event.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after updating event: %w", err)
	}

	if rowsAffected == 0 {
		return ErrEventNotFound
	}

	return nil
}

// Delete removes a group event by its ID
func (r *groupEventRepository) Delete(id string) error {
	query := "DELETE FROM group_events WHERE id = ?"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after deleting event: %w", err)
	}

	if rowsAffected == 0 {
		return ErrEventNotFound
	}

	return nil
}
