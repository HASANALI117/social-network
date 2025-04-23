package services

import (
"errors"
"fmt"
"net/mail"
"time"

"github.com/HASANALI117/social-network/pkg/models"
"github.com/HASANALI117/social-network/pkg/repositories"
"github.com/google/uuid"
"golang.org/x/crypto/bcrypt"
)

// AuthCredentials holds login input
type AuthCredentials struct {
Identifier string // Can be username or email
Password   string
}

var (
ErrInvalidCredentials = errors.New("invalid identifier or password")
)

// AuthService defines the interface for authentication logic
type AuthService interface {
SignIn(credentials AuthCredentials) (session *models.Session, user *UserResponse, err error) // Returns session and sanitized user
SignOut(token string) error
GetUserBySessionToken(token string) (*UserResponse, error) // Replaces helpers.GetUserFromSession
}

// authService implements AuthService interface
type authService struct {
userRepo    repositories.UserRepository
sessionRepo repositories.SessionRepository
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo repositories.UserRepository, sessionRepo repositories.SessionRepository) AuthService {
return &authService{
userRepo:    userRepo,
sessionRepo: sessionRepo,
}
}

// SignIn authenticates a user, creates a session, and returns the session and user details.
func (s *authService) SignIn(credentials AuthCredentials) (*models.Session, *UserResponse, error) {
// 1. Find user by identifier (email or username)
var user *models.User
var err error
// Basic check if identifier looks like an email
if _, parseErr := mail.ParseAddress(credentials.Identifier); parseErr == nil {
user, err = s.userRepo.GetByEmail(credentials.Identifier)
} else {
user, err = s.userRepo.GetByUsername(credentials.Identifier)
}

if err != nil {
if errors.Is(err, repositories.ErrUserNotFound) {
return nil, nil, ErrInvalidCredentials // Don't reveal if user exists
}
return nil, nil, fmt.Errorf("failed to find user: %w", err) // Internal error
}

// 2. Compare password
if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
return nil, nil, ErrInvalidCredentials // Incorrect password
}

// 3. Create session
session := &models.Session{
Token:     uuid.New().String(),
UserID:    user.ID,
ExpiresAt: time.Now().Add(24 * time.Hour), // Example: 24-hour session duration
}

if err := s.sessionRepo.Create(session); err != nil {
return nil, nil, fmt.Errorf("failed to create session: %w", err)
}

// 4. Prepare sanitized user response
userResponse := &UserResponse{
ID:        user.ID,
Username:  user.Username,
Email:     user.Email,
FirstName: user.FirstName,
LastName:  user.LastName,
AvatarURL: user.AvatarURL,
AboutMe:   user.AboutMe,
BirthDate: user.BirthDate,
CreatedAt: user.CreatedAt,
UpdatedAt: user.UpdatedAt,
}

return session, userResponse, nil
}

// SignOut deletes a user session by token.
func (s *authService) SignOut(token string) error {
// DeleteSession in repository handles non-existent tokens gracefully
return s.sessionRepo.DeleteByToken(token)
}

// GetUserBySessionToken validates a session token and returns the associated user details.
func (s *authService) GetUserBySessionToken(token string) (*UserResponse, error) {
// 1. Get session from repository (checks expiry)
session, err := s.sessionRepo.GetByToken(token)
if err != nil {
// Handles ErrSessionNotFound from repo
return nil, err
}

// 2. Get user details using the user ID from the session
user, err := s.userRepo.GetByID(session.UserID)
if err != nil {
// This case (session exists but user doesn't) should ideally not happen
// but handle it defensively. Could indicate data inconsistency.
if errors.Is(err, repositories.ErrUserNotFound) {
// Log this inconsistency
fmt.Printf("Warning: Session token %s valid but user ID %s not found\n", token, session.UserID)
// Invalidate the session as a precaution
_ = s.sessionRepo.DeleteByToken(token)
return nil, repositories.ErrSessionNotFound // Treat as invalid session
}
return nil, fmt.Errorf("failed to get user by ID %s: %w", session.UserID, err)
}

// 3. Prepare sanitized user response
userResponse := &UserResponse{
ID:        user.ID,
Username:  user.Username,
Email:     user.Email,
FirstName: user.FirstName,
LastName:  user.LastName,
AvatarURL: user.AvatarURL,
AboutMe:   user.AboutMe,
BirthDate: user.BirthDate,
CreatedAt: user.CreatedAt,
UpdatedAt: user.UpdatedAt,
}

return userResponse, nil
}
