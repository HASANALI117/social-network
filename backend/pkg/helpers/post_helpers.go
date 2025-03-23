package helpers

import (
	"database/sql"
	"errors"
	"time"

	"social-network/pkg/db"
	"social-network/pkg/models"
	"github.com/google/uuid"
)

var (
	ErrPostNotFound = errors.New("post not found")
)

// type PostDB struct {
// 	db *db.DB
// }

// func NewPostDB(db *db.DB) *PostDB {
// 	return &PostDB{db: db}
// }

func CreatePost(post *models.Post) error {

	query := `
	INSERT INTO posts (id, user_id, title, content, image_url, privacy, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
    `

	post.ID = uuid.New().String()
	post.CreatedAt = time.Now()

	_, err := db.GlobalDB.Exec(query,
		post.ID,
		post.UserID,
		post.Title,
		post.Content,
		post.ImageURL,
		post.Privacy,
		post.CreatedAt,
	)

	return err
}

func GetPostByID(id string) (*models.Post, error) {
	post := &models.Post{}
	query := `
        SELECT id, user_id, title, content, image_url, privacy, created_at
        FROM posts WHERE id = ?
    `

	err := db.GlobalDB.QueryRow(query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.ImageURL,
		&post.Privacy,
		&post.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrPostNotFound
	}
	if err != nil {
		return nil, err
	}

	return post, nil
}

func ListPosts(limit, offset int) ([]*models.Post, error) {
	query := `
        SELECT id, user_id, title, content, image_url, privacy, created_at
        FROM posts
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `

	rows, err := db.GlobalDB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.ImageURL,
			&post.Privacy,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func ListPostsByUser(userID string, limit, offset int) ([]*models.Post, error) {
	query := `
        SELECT id, user_id, title, content, image_url, privacy, created_at
        FROM posts
		WHERE user_id = ?
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `

	rows, err := db.GlobalDB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.ImageURL,
			&post.Privacy,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// func Update(post *models.Post) error {
// 	post.UpdatedAt = time.Now()

// 	query := `
//         UPDATE posts
//         SET title = ?, content = ?, image_url = ?, privacy = ? = ?
//         WHERE id = ?
//     `

// 	result, err := db.GlobalDB.Exec(query,
// 		post.Title,
// 		post.Content,
// 		post.ImageURL,
// 		post.Privacy,
// 		post.UpdatedAt,
// 		post.ID,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	rowsAffected, err := result.RowsAffected()
// 	if err != nil {
// 		return err
// 	}
// 	if rowsAffected == 0 {
// 		return ErrPostNotFound
// 	}

// 	return nil
// }

func DeletePost(id string) error {
	query := `DELETE FROM posts WHERE id = ?`

	result, err := db.GlobalDB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPostNotFound
	}

	return nil
}

func AddPostAllowedUser(postID, userID string) error {
	query := `
        INSERT INTO post_allowed_users (post_id, user_id, created_at)
        VALUES (?, ?, ?)
    `

	_, err := db.GlobalDB.Exec(query, postID, userID, time.Now())
	return err
}

func RemovePostAllowedUser(postID, userID string) error {
	query := `DELETE FROM post_allowed_users WHERE post_id = ? AND user_id = ?`
	_, err := db.GlobalDB.Exec(query, postID, userID)
	return err
}

func GetPostAllowedUsers(postID string) ([]*models.User, error) {
	query := `
        SELECT u.id, u.username, u.email, u.first_name, u.last_name, u.avatar_url, u.about_me, u.birth_date
        FROM users u
        INNER JOIN post_allowed_users pau ON u.id = pau.user_id
        WHERE pau.post_id = ?
    `

	rows, err := db.GlobalDB.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allowedUsers []*models.User
	for rows.Next() {
		allowedUser := &models.User{}
		err := rows.Scan(
			&allowedUser.ID,
			&allowedUser.Username,
			&allowedUser.Email,
			&allowedUser.FirstName,
			&allowedUser.LastName,
			&allowedUser.AvatarURL,
			&allowedUser.AboutMe,
			&allowedUser.BirthDate,
		)
		if err != nil {
			return nil, err
		}
		allowedUsers = append(allowedUsers, allowedUser)
	}

	return allowedUsers, nil
}
