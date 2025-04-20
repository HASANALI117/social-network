package handlers

	"log" // Added for logging potential errors during update
	"net/http"
	"strconv"
	"time" // Added for UserResponse

	"github.com/HASANALI117/social-network/pkg/apperrors" // Added
	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

// --- DTOs (Data Transfer Objects) ---

// RegisterUserRequest defines the expected body for user registration.
type RegisterUserRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	BirthDate string `json:"birth_date"` // Consider using time.Time if format is fixed
	AboutMe   string `json:"about_me"`
	AvatarURL string `json:"avatar_url"` // Optional?
}

// UserResponse defines the user data returned by the API (excluding sensitive info).
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL string    `json:"avatar_url"`
	AboutMe   string    `json:"about_me"`
	BirthDate string    `json:"birth_date"` // Consider time.Time
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListUsersResponse defines the structure for the list users endpoint.
type ListUsersResponse struct {
	Users  []UserResponse `json:"users"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
	Count  int            `json:"count"`
}

// UpdateUserRequest defines the expected body for updating a user.
// Note: Password is included, but logic should handle if it's empty.
type UpdateUserRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password,omitempty"` // Optional password update
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarURL string `json:"avatar_url"`
	AboutMe   string `json:"about_me"`
	BirthDate string `json:"birth_date"`
}

// DeleteUserResponse defines the success message for user deletion.
type DeleteUserResponse struct {
	Message string `json:"message"`
}

// OnlineUser represents a simplified user structure for the online list.
type OnlineUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	// Add other fields if needed, e.g., status
}

// OnlineUsersResponse defines the structure for the online users endpoint.
type OnlineUsersResponse struct {
	OnlineUsers []OnlineUser `json:"online_users"`
	Count       int          `json:"count"`
}

// --- Helper Function ---

// mapModelUserToUserResponse converts a models.User to a UserResponse DTO.
func mapModelUserToUserResponse(user *models.User) UserResponse {
	return UserResponse{
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
}

// --- Handlers ---

// RegisterUser godoc
// @Summary Register a new user
// @Description Register a new user in the system. Requires username, email, and password.
// @Tags users
// @Accept json
// @Produce json
// @Param user body RegisterUserRequest true "User registration details"
// @Success 201 {object} UserResponse "User created successfully"
// @Failure 400 {object} map[string]string "Invalid request body or missing required fields"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to register user (e.g., database error, hashing error)"
// @Router /users/register [post]
func RegisterUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	var req RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apperrors.ErrBadRequest("Invalid request body", err)
	}
	defer r.Body.Close()

	// Basic validation
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return apperrors.ErrBadRequest("Username, email, and password are required", nil)
	}
	// Add more validation as needed (email format, password complexity etc.)

	// Map request DTO to internal model
	user := models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password, // Password will be hashed by CreateUser helper/service
		FirstName: req.FirstName,
		LastName:  req.LastName,
		BirthDate: req.BirthDate,
		AboutMe:   req.AboutMe,
		AvatarURL: req.AvatarURL,
	}

	// Create and Save user to database
	// TODO: Replace helpers.CreateUser when service layer is added
	if err := helpers.CreateUser(&user); err != nil {
		// TODO: Check for specific errors like duplicate username/email if CreateUser provides them
		return apperrors.ErrInternalServer("Failed to register user", err)
	}

	// Return created user DTO
	response := mapModelUserToUserResponse(&user)
	return helpers.RespondJSON(w, http.StatusCreated, response)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get user details by user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id query string true "User ID"
// @Success 200 {object} UserResponse "User details"
// @Failure 400 {object} map[string]string "User ID is required or invalid"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to get user"
// @Router /users/get [get]
func GetUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		return apperrors.ErrBadRequest("User ID is required", nil)
	}
	// Optional: Validate if userID is in the correct format (e.g., UUID)

	// Get user from database
	// TODO: Replace helpers.GetUserByID when service layer is added
	user, err := helpers.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, helpers.ErrUserNotFound) { // Assuming helpers.ErrUserNotFound exists
			return apperrors.ErrNotFound("User not found", err)
		}
		return apperrors.ErrInternalServer("Failed to get user", err)
	}

	// Return user DTO
	response := mapModelUserToUserResponse(user)
	return helpers.RespondJSON(w, http.StatusOK, response)
}

// ListUsers godoc
// @Summary List users
// @Description Get a paginated list of users
// @Tags users
// @Accept json
// @Produce json
// @Param limit query int false "Number of users to return (default 10)" minimum(1) maximum(100)
// @Param offset query int false "Number of users to skip (default 0)" minimum(0)
// @Success 200 {object} ListUsersResponse "List of users"
// @Failure 400 {object} map[string]string "Invalid limit or offset parameter"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to list users"
// @Router /users/list [get]
func ListUsers(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	// Parse pagination parameters with validation
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Default limit
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 || parsedLimit > 100 { // Add max limit
			return apperrors.ErrBadRequest("Invalid limit parameter", err)
		}
		limit = parsedLimit
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil || parsedOffset < 0 {
			return apperrors.ErrBadRequest("Invalid offset parameter", err)
		}
		offset = parsedOffset
	}

	// Get users from database
	// TODO: Replace helpers.ListUsers when service layer is added
	users, err := helpers.ListUsers(limit, offset)
	if err != nil {
		return apperrors.ErrInternalServer("Failed to list users", err)
	}

	// Map models to response DTOs
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = mapModelUserToUserResponse(user)
	}

	// Return user list DTO
	response := ListUsersResponse{
		Users:  userResponses,
		Limit:  limit,
		Offset: offset,
		Count:  len(userResponses), // Note: This might not be the total count in DB
	}
	return helpers.RespondJSON(w, http.StatusOK, response)
}

// UpdateUser godoc
// @Summary Update user
// @Description Update authenticated user's details.
// @Tags users
// @Accept json
// @Produce json
// @Param user body UpdateUserRequest true "User update details (fields to update)"
// @Success 200 {object} UserResponse "Updated user details"
// @Failure 400 {object} map[string]string "Invalid request body or validation error"
// @Failure 401 {object} map[string]string "Unauthorized (user not logged in)"
// @Failure 404 {object} map[string]string "User not found (should not happen if authenticated)"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to update user or hash password"
// @Router /users/update [put]
func UpdateUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPut {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	// Get current user from session for authorization
	// TODO: Replace helpers.GetUserFromSession when auth middleware/service is added
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		// Assuming GetUserFromSession returns an error if not found or invalid
		return apperrors.ErrUnauthorized("Unauthorized", err)
	}

	// Get existing user data (to update)
	// We use currentUser.ID directly, assuming users can only update themselves
	// TODO: Replace helpers.GetUserByID when service layer is added
	user, err := helpers.GetUserByID(currentUser.ID)
	if err != nil {
		if errors.Is(err, helpers.ErrUserNotFound) {
			// This case is unlikely if GetUserFromSession worked, but handle defensively
			log.Printf("ERROR: Authenticated user %s not found in DB for update", currentUser.ID)
			return apperrors.ErrNotFound("Authenticated user not found", err)
		}
		return apperrors.ErrInternalServer("Failed to retrieve user for update", err)
	}

	// Parse request body for updates
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apperrors.ErrBadRequest("Invalid request body", err)
	}
	defer r.Body.Close()

	// Apply updates from request DTO to the model
	// Only update fields if they are provided in the request (or handle defaults)
	// Basic validation: Ensure required fields aren't being cleared if they shouldn't be.
	if req.Username == "" || req.Email == "" {
		return apperrors.ErrBadRequest("Username and email cannot be empty", nil)
	}
	user.Username = req.Username
	user.Email = req.Email
	user.FirstName = req.FirstName // Allow clearing these? Add validation if not.
	user.LastName = req.LastName
	user.AvatarURL = req.AvatarURL
	user.AboutMe = req.AboutMe
	user.BirthDate = req.BirthDate

	// Only update password if provided in the request
	if req.Password != "" {
		// TODO: Move password hashing to CreateUser/UpdateUser helper/service
		hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if hashErr != nil {
			return apperrors.ErrInternalServer("Failed to hash new password", hashErr)
		}
		user.Password = string(hashedPassword) // Update the model's password field
	} else {
		// Ensure the existing password hash isn't overwritten if not provided
		// GetUserByID should have populated user.Password already.
	}

	// Save updated user to database
	// TODO: Replace helpers.UpdateUser when service layer is added
	if err := helpers.UpdateUser(user); err != nil {
		// TODO: Handle potential specific errors from UpdateUser (e.g., duplicate email/username on update)
		return apperrors.ErrInternalServer("Failed to update user", err)
	}

	// Fetch the updated user again to get the latest UpdatedAt timestamp
	// Alternatively, UpdateUser could return the updated user model.
	updatedUser, err := helpers.GetUserByID(currentUser.ID)
	if err != nil {
		log.Printf("WARN: Failed to fetch user after update, returning potentially stale data: %v", err)
		// Fallback to returning the user model we have, but UpdatedAt might be old
		updatedUser = user
	}


	// Return updated user DTO
	response := mapModelUserToUserResponse(updatedUser)
	return helpers.RespondJSON(w, http.StatusOK, response)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete a user by ID. Requires appropriate authorization (e.g., admin or self).
// @Tags users
// @Accept json
// @Produce json
// @Param id query string true "User ID to delete"
// @Success 200 {object} DeleteUserResponse "User deleted successfully"
// @Failure 400 {object} map[string]string "User ID is required"
// @Failure 401 {object} map[string]string "Unauthorized" // Added for auth check
// @Failure 403 {object} map[string]string "Forbidden" // Added if user tries to delete someone else
// @Failure 404 {object} map[string]string "User not found"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to delete user"
// @Router /users/delete [delete]
func DeleteUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodDelete {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	// Get user ID to delete from URL query parameter
	userIDToDelete := r.URL.Query().Get("id")
	if userIDToDelete == "" {
		return apperrors.ErrBadRequest("User ID is required", nil)
	}

	// --- Authorization Check ---
	// Get current user from session
	// TODO: Replace helpers.GetUserFromSession when auth middleware/service is added
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		return apperrors.ErrUnauthorized("Unauthorized", err)
	}

	// Basic authorization: Allow users to delete themselves, or implement admin logic
	if currentUser.ID != userIDToDelete {
		// TODO: Add proper role-based authorization check here if needed
		log.Printf("WARN: User %s attempted to delete user %s", currentUser.ID, userIDToDelete)
		return apperrors.ErrForbidden("You are not allowed to delete this user", nil)
	}
	// --- End Authorization Check ---


	// Delete user from database
	// TODO: Replace helpers.DeleteUser when service layer is added
	if err := helpers.DeleteUser(userIDToDelete); err != nil {
		if errors.Is(err, helpers.ErrUserNotFound) { // Assuming helpers.ErrUserNotFound exists
			return apperrors.ErrNotFound("User not found", err)
		}
		return apperrors.ErrInternalServer("Failed to delete user", err)
	}

	// Return success response
	response := DeleteUserResponse{Message: "User deleted successfully"}
	return helpers.RespondJSON(w, http.StatusOK, response)
}

// OnlineUsers godoc
// @Summary Get online users
// @Description Get a list of currently online users (excluding the requesting user). Requires authentication.
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} OnlineUsersResponse "List of online users"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Internal server error (e.g., fetching session user)"
// @Router /users/online [get]
func OnlineUsers(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	// Get current user from session
	// TODO: Replace helpers.GetUserFromSession when auth middleware/service is added
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		return apperrors.ErrUnauthorized("Unauthorized", err)
	}

	// Get online users from WebSocket hub
	// TODO: Ensure WebSocketHub is properly managed and injected if needed, or accessed safely
	onlineUsersMap := WebSocketHub.GetUsersWithStatus() // Assuming this returns []map[string]string

	// Filter out current user and map to DTO
	filteredUsers := make([]OnlineUser, 0, len(onlineUsersMap))
	for _, userMap := range onlineUsersMap {
		userID := userMap["id"] // Assuming keys are "id" and "username"
		if userID != currentUser.ID {
			filteredUsers = append(filteredUsers, OnlineUser{
				ID:       userID,
				Username: userMap["username"],
			})
		}
	}

	// Return the list DTO
	response := OnlineUsersResponse{
		OnlineUsers: filteredUsers,
		Count:       len(filteredUsers),
	}
	return helpers.RespondJSON(w, http.StatusOK, response)
}
