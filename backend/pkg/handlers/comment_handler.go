package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"social-network/pkg/models"

	"github.com/google/uuid"
)

func CreateComment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var comment models.Comment
		if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate input
		if comment.PostID == "" || comment.Content == "" {
			http.Error(w, "post_id and content are required", http.StatusBadRequest)
			return
		}

		// Assume user_id comes from authentication middleware (e.g., session or token)
		userID := "authenticated-user-id" // Replace with actual logic
		comment.ID = uuid.New().String()
		comment.UserID = userID

		// Insert into database
		query := `INSERT INTO comments (id, post_id, user_id, content) VALUES (?, ?, ?, ?)`
		_, err := db.Exec(query, comment.ID, comment.PostID, comment.UserID, comment.Content)
		if err != nil {
			http.Error(w, "Failed to create comment", http.StatusInternalServerError)
			return
		}

		// Return the created comment
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(comment)
	}
}

func GetCommentsByPost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID := r.URL.Query().Get("postId")
		if postID == "" {
			http.Error(w, "postId query parameter is required", http.StatusBadRequest)
			return
		}

		// Query comments from database
		query := `SELECT id, post_id, user_id, content, created_at FROM comments WHERE post_id = ?`
		rows, err := db.Query(query, postID)
		if err != nil {
			http.Error(w, "Failed to retrieve comments", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var comments []models.Comment
		for rows.Next() {
			var comment models.Comment
			if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt); err != nil {
				http.Error(w, "Failed to process comments", http.StatusInternalServerError)
				return
			}
			comments = append(comments, comment)
		}

		// Return the comments
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comments)
	}
}
