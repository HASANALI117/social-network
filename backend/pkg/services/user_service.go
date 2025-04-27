package services

import (
	"fmt"
	"time"

	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/repositories"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserResponse defines the sanitized user data returned by service methods
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL string    `json:"avatar_url"`
	AboutMe   string    `json:"about_me"`
	BirthDate string    `json:"birth_date"`
	IsPrivate bool      `json:"is_private"` // Added privacy status
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserService defines the interface for user business logic
type UserService interface {
	Register(user *models.User) (*UserResponse, error)
	GetByID(id string) (*UserResponse, error)
	GetByUsername(username string) (*UserResponse, error)
	GetByEmail(email string) (*UserResponse, error)
	Update(id string, updateData map[string]interface{}) (*UserResponse, error)
	Delete(id string) error
	List(limit, offset int) ([]*UserResponse, error)
	UpdatePrivacy(userID string, isPrivate bool) error // Added method
}

// userService implements UserService interface
type userService struct {
	userRepo repositories.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) Register(user *models.User) (*UserResponse, error) {
	// Validate input
	if user.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if user.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if user.Password == "" {
		return nil, fmt.Errorf("password is required")
	}
	if len(user.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}
	// Generate UUID for new user
	user.ID = uuid.New().String()

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err) // Return nil UserResponse on error
	}
	user.Password = string(hashedPassword)

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Create user in repository
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Return sanitized response
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarURL: user.AvatarURL,
		AboutMe:   user.AboutMe,
		BirthDate: user.BirthDate,
		IsPrivate: user.IsPrivate, // Added privacy status
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *userService) GetByID(id string) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarURL: user.AvatarURL,
		AboutMe:   user.AboutMe,
		BirthDate: user.BirthDate,
		IsPrivate: user.IsPrivate, // Added privacy status
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *userService) GetByUsername(username string) (*UserResponse, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarURL: user.AvatarURL,
		AboutMe:   user.AboutMe,
		BirthDate: user.BirthDate,
		IsPrivate: user.IsPrivate, // Added privacy status
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *userService) GetByEmail(email string) (*UserResponse, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarURL: user.AvatarURL,
		AboutMe:   user.AboutMe,
		BirthDate: user.BirthDate,
		IsPrivate: user.IsPrivate, // Added privacy status
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *userService) Update(id string, updateData map[string]interface{}) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err // Handles ErrUserNotFound from repo
	}

	// Update fields based on updateData map
	if username, ok := updateData["username"].(string); ok {
		user.Username = username
	}
	if email, ok := updateData["email"].(string); ok {
		user.Email = email
	}
	if firstName, ok := updateData["first_name"].(string); ok {
		user.FirstName = firstName
	}
	if lastName, ok := updateData["last_name"].(string); ok {
		user.LastName = lastName
	}
	if avatarURL, ok := updateData["avatar_url"].(string); ok {
		user.AvatarURL = avatarURL
	}
	if aboutMe, ok := updateData["about_me"].(string); ok {
		user.AboutMe = aboutMe
	}
	if birthDate, ok := updateData["birth_date"].(string); ok {
		user.BirthDate = birthDate
	}

	// Handle password update separately
	if password, ok := updateData["password"].(string); ok && password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password during update: %w", err)
		}
		user.Password = string(hashedPassword)
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Save updated user
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	// Return sanitized response
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarURL: user.AvatarURL,
		AboutMe:   user.AboutMe,
		BirthDate: user.BirthDate,
		IsPrivate: user.IsPrivate, // Added privacy status
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *userService) Delete(id string) error {
	return s.userRepo.Delete(id)
}

func (s *userService) List(limit, offset int) ([]*UserResponse, error) {
	users, err := s.userRepo.List(limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []*UserResponse
	for _, user := range users {
		responses = append(responses, &UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			AvatarURL: user.AvatarURL,
			AboutMe:   user.AboutMe,
			BirthDate: user.BirthDate,
			IsPrivate: user.IsPrivate, // Added privacy status
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}
	return responses, nil
}

// UpdatePrivacy updates the privacy setting for a user
// Note: Assumes authorization (checking if the caller can update this user's privacy)
// is handled before calling this service method, likely in the handler.
func (s *userService) UpdatePrivacy(userID string, isPrivate bool) error {
	// Optionally, re-fetch the user here if needed for other checks,
	// but the repository method handles the update directly.
	err := s.userRepo.UpdatePrivacy(userID, isPrivate)
	if err != nil {
		// Wrap or handle specific errors like ErrUserNotFound if necessary
		return fmt.Errorf("failed to update user privacy in service: %w", err)
	}
	return nil
}
