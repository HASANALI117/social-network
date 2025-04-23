package services

import (
	"fmt"
	"time"

	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/repositories"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserService defines the interface for user business logic
type UserService interface {
	Register(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(id string, updateData map[string]interface{}) (*models.User, error)
	Delete(id string) error
	List(limit, offset int) ([]*models.User, error)
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

func (s *userService) Register(user *models.User) error {
	// Generate UUID for new user
	user.ID = uuid.New().String()

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Create user in repository
	if err := s.userRepo.Create(user); err != nil {
		// Handle potential duplicate errors if necessary
		return err
	}

	return nil
}

func (s *userService) GetByID(id string) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *userService) GetByUsername(username string) (*models.User, error) {
	return s.userRepo.GetByUsername(username)
}

func (s *userService) GetByEmail(email string) (*models.User, error) {
	return s.userRepo.GetByEmail(email)
}

func (s *userService) Update(id string, updateData map[string]interface{}) (*models.User, error) {
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

	return user, nil
}

func (s *userService) Delete(id string) error {
	return s.userRepo.Delete(id)
}

func (s *userService) List(limit, offset int) ([]*models.User, error) {
	return s.userRepo.List(limit, offset)
}
