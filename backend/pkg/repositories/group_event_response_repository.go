package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	// Needed for error checking
	// Needed for time.Now()
	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/google/uuid"
	// Needed for uuid.New()
)

// GroupEventResponseRepository defines the interface for accessing group event response data
type GroupEventResponseRepository interface {
	// CreateOrUpdate inserts a new response or updates the existing one for the same user and event.
	CreateOrUpdate(response *models.GroupEventResponse) error
	// GetByEventID retrieves all responses for a specific event.
	GetByEventID(eventID string) ([]*models.GroupEventResponse, error)
	// GetCountsByEventID retrieves the count of 'going' and 'not_going' responses for an event.
	GetCountsByEventID(eventID string) (goingCount int, notGoingCount int, err error)
	// TODO: Add GetByEventAndUser if needed later
}

// groupEventResponseRepository implements GroupEventResponseRepository
type groupEventResponseRepository struct {
	db *sql.DB
}

// NewGroupEventResponseRepository creates a new GroupEventResponseRepository
func NewGroupEventResponseRepository(db *sql.DB) GroupEventResponseRepository {
	return &groupEventResponseRepository{
		db: db,
	}
}

// --- Method Implementations (To be added) ---

// CreateOrUpdate inserts a new response or updates the existing one.
// Uses SQLite's UPSERT functionality (INSERT ... ON CONFLICT ... DO UPDATE).
func (r *groupEventResponseRepository) CreateOrUpdate(response *models.GroupEventResponse) error {
	query := `
        INSERT INTO group_event_responses (id, event_id, user_id, response, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?)
        ON CONFLICT(event_id, user_id) DO UPDATE SET
            response = excluded.response,
            updated_at = ?
    `
	now := time.Now()
	newUUID := uuid.New().String()

	// Prepare arguments for both INSERT and UPDATE cases
	// Note: SQLite's excluded.column refers to the value that *would* have been inserted.
	// We need to provide the updated_at value separately for the UPDATE clause.
	_, err := r.db.Exec(
		query,
		newUUID, // id for potential INSERT
		response.EventID,
		response.UserID,
		response.Response, // response for potential INSERT
		now,               // created_at for potential INSERT
		now,               // updated_at for potential INSERT
		// --- Arguments for ON CONFLICT DO UPDATE ---
		// response = excluded.response is handled by SQLite
		now, // updated_at for UPDATE
	)

	if err != nil {
		// Check for foreign key violation (event or user doesn't exist)
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return fmt.Errorf("failed to create/update response: event or user not found: %w", err)
		}
		// Check for CHECK constraint violation (invalid response value)
		if strings.Contains(err.Error(), "CHECK constraint failed") {
			return fmt.Errorf("failed to create/update response: invalid response value '%s': %w", response.Response, err)
		}
		return fmt.Errorf("failed to create/update group event response: %w", err)
	}

	// Since we don't know if it was an insert or update easily without another query,
	// we don't explicitly set the ID or timestamps back on the input 'response' object here.
	// The caller should ideally refetch if they need the guaranteed latest state including ID/timestamps.
	return nil
}

// GetByEventID retrieves all responses for a specific event.
func (r *groupEventResponseRepository) GetByEventID(eventID string) ([]*models.GroupEventResponse, error) {
	query := `
        SELECT id, event_id, user_id, response, created_at, updated_at
        FROM group_event_responses
        WHERE event_id = ?
        ORDER BY updated_at DESC
    `
	rows, err := r.db.Query(query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to query group event responses by event ID: %w", err)
	}
	defer rows.Close()

	responses := make([]*models.GroupEventResponse, 0)
	for rows.Next() {
		var resp models.GroupEventResponse
		var createdAtStr, updatedAtStr string // Scan into strings first

		err := rows.Scan(
			&resp.ID,
			&resp.EventID,
			&resp.UserID,
			&resp.Response,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group event response row: %w", err)
		}

		// Parse timestamps
		resp.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			// Log or handle parsing error - maybe return partial results or error out?
			fmt.Printf("Warning: Failed to parse created_at timestamp '%s' for response %s: %v\n", createdAtStr, resp.ID, err)
		}
		resp.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			fmt.Printf("Warning: Failed to parse updated_at timestamp '%s' for response %s: %v\n", updatedAtStr, resp.ID, err)
		}

		responses = append(responses, &resp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating group event response rows: %w", err)
	}

	return responses, nil
}

// GetCountsByEventID retrieves the count of 'going' and 'not_going' responses.
func (r *groupEventResponseRepository) GetCountsByEventID(eventID string) (goingCount int, notGoingCount int, err error) {
	query := `
        SELECT
            SUM(CASE WHEN response = 'going' THEN 1 ELSE 0 END) as going_count,
            SUM(CASE WHEN response = 'not_going' THEN 1 ELSE 0 END) as not_going_count
        FROM group_event_responses
        WHERE event_id = ?
    `
	// Use QueryRow because we expect exactly one row (even if counts are 0)
	// Need to handle potential NULL results if no rows match the WHERE clause, though SUM should return 0.
	// Using sql.NullInt64 for scanning to handle potential NULLs safely.
	var nullGoing sql.NullInt64
	var nullNotGoing sql.NullInt64

	err = r.db.QueryRow(query, eventID).Scan(&nullGoing, &nullNotGoing)
	if err != nil {
		// sql.ErrNoRows should not happen with SUM, but handle defensively
		if err == sql.ErrNoRows {
			return 0, 0, nil // No responses found for this event, counts are zero
		}
		return 0, 0, fmt.Errorf("failed to query group event response counts: %w", err)
	}

	// Assign counts from NullInt64
	goingCount = int(nullGoing.Int64)       // Defaults to 0 if nullGoing.Valid is false
	notGoingCount = int(nullNotGoing.Int64) // Defaults to 0 if nullNotGoing.Valid is false

	return goingCount, notGoingCount, nil
}
