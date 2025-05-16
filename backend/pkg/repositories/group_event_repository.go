package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/types" // Added import
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
	GetEventsByGroupID(groupID string, upcomingOnly bool) ([]types.EventSummary, error)
	GetEventWithResponsesByID(eventID string) (*models.GroupEventAPI, error) // New method
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

// GetEventsByGroupID retrieves a list of event summaries for a specific group.
// If upcomingOnly is true, it only returns events with event_time in the future.
func (r *groupEventRepository) GetEventsByGroupID(groupID string, upcomingOnly bool) ([]types.EventSummary, error) {
	baseQuery := `
		SELECT
			e.id AS event_id,
			e.title,
			SUBSTR(e.description, 1, 100) AS description_snippet, -- Adjust snippet length
			e.event_time
			-- e.location -- Removed as its existence in DB is uncertain, causes query to fail if column missing
		FROM group_events e
		WHERE e.group_id = ?
	`
	args := []interface{}{groupID}

	if upcomingOnly {
		baseQuery += " AND e.event_time >= ?"
		args = append(args, time.Now().UTC())
	}

	baseQuery += " ORDER BY e.event_time ASC" // Or DESC for most recent upcoming

	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list event summaries by group ID %s: %w", groupID, err)
	}
	defer rows.Close()

	summaries := make([]types.EventSummary, 0)
	for rows.Next() {
		var summary types.EventSummary
		var eventTimeStr string
		// var location sql.NullString // Removed
		var descriptionSnippet sql.NullString

		err := rows.Scan(
			&summary.EventID,
			&summary.Title,
			&descriptionSnippet,
			&eventTimeStr,
			// &location, // Removed
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event summary for group ID %s: %w", groupID, err)
		}

		if descriptionSnippet.Valid {
			summary.DescriptionSnippet = descriptionSnippet.String
		}
		// if location.Valid { // Removed
		// 	summary.Location = location.String // Removed
		// }

		summary.StartTime, err = time.Parse(time.RFC3339, eventTimeStr)
		if err != nil {
			// Attempt to parse with "YYYY-MM-DD HH:MM:SS" if RFC3339 fails, as SQLite might store it this way
			parsedTime, errFallback := time.Parse("2006-01-02 15:04:05", eventTimeStr)
			if errFallback != nil {
				fmt.Printf("Warning: Failed to parse event_time timestamp '%s' for event summary %s (tried RFC3339 and YYYY-MM-DD HH:MM:SS): %v\n", eventTimeStr, summary.EventID, err) // Log original error
				summary.StartTime = time.Time{}
			} else {
				summary.StartTime = parsedTime
			}
		}
		summaries = append(summaries, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event summary list for group ID %s rows: %w", groupID, err)
	}

	return summaries, nil
}

// GetEventWithResponsesByID retrieves a group event by its ID, along with its responses enriched with user details.
func (r *groupEventRepository) GetEventWithResponsesByID(eventID string) (*models.GroupEventAPI, error) {
	// 1. Fetch the main event details
	eventQuery := `
		SELECT
			e.id, e.group_id, e.creator_id, e.title, e.description, e.event_time, e.created_at, e.updated_at,
			g.avatar_url AS group_avatar_url
		FROM group_events e
		JOIN groups g ON e.group_id = g.id
		WHERE e.id = ?
	`
	var event models.GroupEvent
	err := r.db.QueryRow(eventQuery, eventID).Scan(
		&event.ID,
		&event.GroupID,
		&event.CreatorID,
		&event.Title,
		&event.Description,
		&event.EventTime,
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.GroupAvatarURL,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get event by ID for GetEventWithResponsesByID: %w", err)
	}

	eventAPI := &models.GroupEventAPI{
		GroupEvent: event,
		Responses:  []models.EventResponseAPI{}, // Initialize as empty slice
	}

	// 2. Fetch the event responses with user details
	responsesQuery := `
	       SELECT
	           ger.user_id,
	           u.username,
	           u.first_name,
	           u.last_name,
	           u.avatar_url,
	           ger.response,
	           ger.updated_at
	       FROM group_event_responses ger
	       JOIN users u ON ger.user_id = u.id
	       WHERE ger.event_id = ?
	       ORDER BY ger.updated_at DESC
	   `
	rows, err := r.db.Query(responsesQuery, eventID)
	if err != nil {
		// If error is sql.ErrNoRows, it means no responses, which is fine.
		// For other errors, return the error.
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to query event responses for GetEventWithResponsesByID: %w", err)
		}
		// No responses found, return event with empty responses list
		return eventAPI, nil
	}
	defer rows.Close()

	for rows.Next() {
		var resp models.EventResponseAPI
		var firstName, lastName, avatarURL sql.NullString // Handle nullable fields
		var updatedAtStr string                          // Assuming updated_at from DB is string

		err := rows.Scan(
			&resp.UserID,
			&resp.Username,
			&firstName,
			&lastName,
			&avatarURL,
			&resp.Response,
			&updatedAtStr, // Scan as string first
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event response for GetEventWithResponsesByID: %w", err)
		}

		if firstName.Valid {
			resp.FirstName = firstName.String
		}
		if lastName.Valid {
			resp.LastName = lastName.String
		}
		if avatarURL.Valid {
			resp.AvatarURL = avatarURL.String
		}
		resp.UpdatedAt = updatedAtStr // Assign string directly as per model

		eventAPI.Responses = append(eventAPI.Responses, resp)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event response rows for GetEventWithResponsesByID: %w", err)
	}

	return eventAPI, nil
}
