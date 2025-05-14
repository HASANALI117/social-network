package services

import (
	"database/sql" // Added for sql.ErrNoRows
	"errors"       // Added for errors.Is
	"fmt"
	"log" // Added for logging
	"time"

	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/repositories"
	"github.com/HASANALI117/social-network/pkg/types"
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

// UserProfileResponse defines the detailed user profile data including activity
type UserProfileResponse struct {
	UserResponse                       // Embed basic user info
	FollowersCount     int             `json:"followers_count"` // Added follower count
	FollowingCount     int             `json:"following_count"` // Added following count
	IsFollowed         bool            `json:"is_followed"`
	FollowRequestState string          `json:"follow_request_state,omitempty"`
	LatestPosts        []*PostResponse `json:"latest_posts"`     // Changed to []*PostResponse
	LatestFollowers    []models.User   `json:"latest_followers"` // Add followers
	LatestFollowing    []models.User   `json:"latest_following"` // Add following
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
	UpdatePrivacy(userID string, isPrivate bool) error                           // Added method
	GetUserProfile(viewerID, profileUserID string) (*UserProfileResponse, error) // Added method
	SearchUsers(query string) ([]types.UserSearchResultDTO, error)
}

// userService implements UserService interface
type userService struct {
	userRepo        repositories.UserRepository
	postService     PostService     // Added PostService dependency
	followerService FollowerService // Added FollowerService dependency
}

// NewUserService creates a new UserService
func NewUserService(userRepo repositories.UserRepository, postService PostService, followerService FollowerService) UserService {
	return &userService{
		userRepo:        userRepo,
		postService:     postService,
		followerService: followerService,
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

// ErrForbidden is returned when a user is not allowed to access a resource.
var ErrForbidden = errors.New("access forbidden")

// GetUserProfile retrieves detailed profile information, respecting privacy settings.
func (s *userService) GetUserProfile(viewerID, profileUserID string) (*UserProfileResponse, error) {
	const profileDataLimit = 10 // Limit for posts, followers, following

	// 1. Get the target profile user
	profileUser, err := s.userRepo.GetByID(profileUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, repositories.ErrUserNotFound) {
			return nil, repositories.ErrUserNotFound
		}
		log.Printf("Error getting profile user %s: %v", profileUserID, err)
		return nil, fmt.Errorf("internal server error retrieving profile user")
	}

	// 2. Determine viewer's relationship with the profile user
	isSelf := viewerID == profileUserID
	isFollowedByViewer := false       // True if viewer has an "accepted" follow to profileUser
	effectiveFollowRequestState := "" // Will be "SENT" (viewer to profile), "RECEIVED" (profile to viewer), or ""

	if viewerID != "" && !isSelf {
		// Check viewer's follow status towards profileUser (outgoing)
		outgoingFollow, err := s.followerService.FindFollow(viewerID, profileUserID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("Error checking outgoing follow status from %s to %s: %v", viewerID, profileUserID, err)
			// Consider if this error should halt the process or be logged and ignored for status calculation
		}
		if outgoingFollow != nil {
			if outgoingFollow.Status == "accepted" {
				isFollowedByViewer = true
			} else if outgoingFollow.Status == "pending" {
				effectiveFollowRequestState = "SENT"
			}
		}

		// If viewer hasn't sent a pending request to profileUser,
		// check if profileUser has sent a pending request to viewer (incoming for viewer)
		// This determines the "RECEIVED" state.
		if effectiveFollowRequestState == "" { // Only if not "SENT"
			incomingFollow, err := s.followerService.FindFollow(profileUserID, viewerID) // profileUser is follower, viewerID is target
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				log.Printf("Error checking incoming follow status to %s from %s: %v", viewerID, profileUserID, err)
			}
			if incomingFollow != nil && incomingFollow.Status == "pending" {
				effectiveFollowRequestState = "RECEIVED"
			}
		}
	}

	// 3. Determine if the viewer has full access to the profile
	canViewFullProfile := !profileUser.IsPrivate || isSelf || isFollowedByViewer

	// 4. Construct and return response based on access level
	if !canViewFullProfile {
		// This branch is for: private profile, viewer is not self, viewer is not following.
		// Minimal profile response.
		// For minimal view, FollowRequestState is only "SENT" if viewer sent a pending request, otherwise empty.
		minimalFollowRequestState := ""
		if effectiveFollowRequestState == "SENT" {
			minimalFollowRequestState = effectiveFollowRequestState
		}

		return &UserProfileResponse{
			UserResponse: UserResponse{
				ID:        profileUser.ID,
				Username:  profileUser.Username,
				FirstName: profileUser.FirstName,
				LastName:  profileUser.LastName,
				AvatarURL: profileUser.AvatarURL,
				IsPrivate: profileUser.IsPrivate, // This will be true in this scenario
			},
			IsFollowed:         false,
			FollowRequestState: minimalFollowRequestState,
		}, nil
	}

	// --- Full Profile Response ---
	// Viewer has full access (public profile, or private profile viewed by self or an accepted follower)

	var postsResponse []*PostResponse
	var followers []models.User
	var following []models.User
	var followersCount int
	var followingCount int
	var fetchErr error

	followersCount, fetchErr = s.followerService.CountFollowers(profileUserID)
	if fetchErr != nil {
		log.Printf("Error fetching followers count for user %s: %v", profileUserID, fetchErr)
		followersCount = 0
	}

	followingCount, fetchErr = s.followerService.CountFollowing(profileUserID)
	if fetchErr != nil {
		log.Printf("Error fetching following count for user %s: %v", profileUserID, fetchErr)
		followingCount = 0
	}

	postsResponse, fetchErr = s.postService.ListPostsByUser(profileUserID, viewerID, profileDataLimit, 0)
	if fetchErr != nil {
		log.Printf("Error fetching posts for user %s (viewer %s): %v", profileUserID, viewerID, fetchErr)
		postsResponse = []*PostResponse{}
	}

	followers, fetchErr = s.followerService.ListFollowers(profileUserID, profileDataLimit, 0)
	if fetchErr != nil {
		log.Printf("Error fetching followers for user %s: %v", profileUserID, fetchErr)
		followers = []models.User{}
	}

	following, fetchErr = s.followerService.ListFollowing(profileUserID, profileDataLimit, 0)
	if fetchErr != nil {
		log.Printf("Error fetching following for user %s: %v", profileUserID, fetchErr)
		following = []models.User{}
	}

	fullUserResponseData := UserResponse{
		ID:        profileUser.ID,
		Username:  profileUser.Username,
		Email:     profileUser.Email,
		FirstName: profileUser.FirstName,
		LastName:  profileUser.LastName,
		AvatarURL: profileUser.AvatarURL,
		AboutMe:   profileUser.AboutMe,
		BirthDate: profileUser.BirthDate,
		IsPrivate: profileUser.IsPrivate,
		CreatedAt: profileUser.CreatedAt,
		UpdatedAt: profileUser.UpdatedAt,
	}

	return &UserProfileResponse{
		UserResponse:       fullUserResponseData,
		FollowersCount:     followersCount,
		FollowingCount:     followingCount,
		IsFollowed:         isFollowedByViewer,
		FollowRequestState: effectiveFollowRequestState,
		LatestPosts:        postsResponse,
		LatestFollowers:    followers,
		LatestFollowing:    following,
	}, nil
}

// SearchUsers searches for users based on a query string.
func (s *userService) SearchUsers(query string) ([]types.UserSearchResultDTO, error) {
	// Basic validation
	if query == "" {
		return []types.UserSearchResultDTO{}, nil // Return empty if query is empty, or an error
	}

	// Define a reasonable limit for search results
	const searchLimit = 20 // This could be configurable

	users, err := s.userRepo.SearchUsers(query, searchLimit)
	if err != nil {
		// Log the error for internal tracking
		log.Printf("Error searching users with query '%s': %v", query, err)
		// Return a generic error to the caller
		return nil, fmt.Errorf("failed to search users")
	}

	return users, nil
}
