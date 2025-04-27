package services

import (
	"database/sql" // Added for sql.ErrNoRows
	"errors"
	"fmt"
	"log"

	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/repositories"
)

// FollowerService defines the interface for follower business logic
type FollowerService interface {
	RequestFollow(requesterID, targetID string) error
	AcceptFollow(accepterID, requesterID string) error
	RejectFollow(rejecterID, requesterID string) error
	Unfollow(unfollowerID, targetID string) error
	ListFollowers(userID string, limit, offset int) ([]models.User, error) // Added pagination
	ListFollowing(userID string, limit, offset int) ([]models.User, error) // Added pagination
	ListPendingRequests(userID string) (map[string][]models.User, error)   // Changed return type
}

// followerService implements FollowerService
type followerService struct {
	followerRepo repositories.FollowerRepository
	userRepo     repositories.UserRepository // Assuming UserRepository exists and is needed
}

// NewFollowerService creates a new instance of FollowerService
func NewFollowerService(followerRepo repositories.FollowerRepository, userRepo repositories.UserRepository) FollowerService {
	return &followerService{
		followerRepo: followerRepo,
		userRepo:     userRepo,
	}
}

// RequestFollow handles the logic for sending a follow request
func (s *followerService) RequestFollow(requesterID, targetID string) error {
	if requesterID == targetID {
		return errors.New("cannot follow yourself")
	}

	// Check if target user exists and get their privacy status
	targetUser, err := s.userRepo.GetByID(targetID) // Corrected method name
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, repositories.ErrUserNotFound) { // Check for specific not found errors
			return errors.New("target user not found")
		}
		log.Printf("Error checking target user: %v", err)
		return fmt.Errorf("internal server error checking target user")
	}

	// Check if already following or request pending
	existing, err := s.followerRepo.FindFollow(requesterID, targetID)
	if err != nil {
		log.Printf("Error checking existing follow: %v", err)
		return fmt.Errorf("internal server error checking follow status")
	}
	if existing != nil {
		if existing.Status == "accepted" {
			return errors.New("already following this user")
		}
		if existing.Status == "pending" {
			return errors.New("follow request already pending")
		}
		// Handle unexpected status if necessary
		return fmt.Errorf("unexpected follow status: %s", existing.Status)
	}

	// Create the follow request/relationship based on privacy
	if targetUser.IsPrivate {
		// Private profile: Create a pending request
		err = s.followerRepo.CreateFollowRequest(requesterID, targetID)
		if err != nil {
			log.Printf("Error creating follow request for private profile: %v", err)
			return fmt.Errorf("failed to send follow request")
		}
		log.Printf("Follow request sent from %s to private user %s", requesterID, targetID)
	} else {
		// Public profile: Create request and immediately accept it
		err = s.followerRepo.CreateFollowRequest(requesterID, targetID)
		if err != nil {
			// Handle potential duplicate error if CreateFollowRequest fails uniquely
			log.Printf("Error creating initial follow record for public profile: %v", err)
			return fmt.Errorf("failed to initiate follow for public profile")
		}
		err = s.followerRepo.UpdateFollowStatus(requesterID, targetID, "accepted")
		if err != nil {
			log.Printf("Error auto-accepting follow for public profile: %v", err)
			// Consider cleanup: Delete the pending request if acceptance fails?
			// s.followerRepo.DeleteFollow(requesterID, targetID) // Optional cleanup
			return fmt.Errorf("failed to finalize follow for public profile")
		}
		log.Printf("User %s automatically followed public user %s", requesterID, targetID)
	}

	return nil
}

// AcceptFollow handles the logic for accepting a follow request
func (s *followerService) AcceptFollow(accepterID, requesterID string) error {
	// Check if a pending request exists from requester to accepter
	request, err := s.followerRepo.FindFollow(requesterID, accepterID)
	if err != nil {
		log.Printf("Error finding follow request to accept: %v", err)
		return fmt.Errorf("internal server error checking follow request")
	}
	if request == nil || request.Status != "pending" {
		return errors.New("no pending follow request found from this user")
	}

	// Update status to accepted
	err = s.followerRepo.UpdateFollowStatus(requesterID, accepterID, "accepted")
	if err != nil {
		log.Printf("Error accepting follow request: %v", err)
		return fmt.Errorf("failed to accept follow request")
	}

	return nil
}

// RejectFollow handles the logic for rejecting or deleting a follow request
func (s *followerService) RejectFollow(rejecterID, requesterID string) error {
	// Check if a pending request exists from requester to rejecter
	request, err := s.followerRepo.FindFollow(requesterID, rejecterID)
	if err != nil {
		log.Printf("Error finding follow request to reject: %v", err)
		return fmt.Errorf("internal server error checking follow request")
	}
	if request == nil {
		return errors.New("no follow request found from this user to reject")
	}
	// It might be okay to reject an already accepted follow, effectively deleting it.
	// Or enforce only rejecting pending requests:
	// if request.Status != "pending" {
	//     return errors.New("cannot reject a non-pending request")
	// }

	// Delete the follow record (whether pending or accepted)
	err = s.followerRepo.DeleteFollow(requesterID, rejecterID)
	if err != nil {
		log.Printf("Error rejecting/deleting follow request: %v", err)
		return fmt.Errorf("failed to reject follow request")
	}

	return nil
}

// Unfollow handles the logic for removing an accepted follow relationship
func (s *followerService) Unfollow(unfollowerID, targetID string) error {
	// Check if currently following
	follow, err := s.followerRepo.FindFollow(unfollowerID, targetID)
	if err != nil {
		log.Printf("Error finding follow to unfollow: %v", err)
		return fmt.Errorf("internal server error checking follow status")
	}
	if follow == nil || follow.Status != "accepted" {
		return errors.New("not following this user")
	}

	// Delete the follow record
	err = s.followerRepo.DeleteFollow(unfollowerID, targetID)
	if err != nil {
		log.Printf("Error unfollowing user: %v", err)
		return fmt.Errorf("failed to unfollow user")
	}

	return nil
}

// ListFollowers retrieves a paginated list of users following the given userID
func (s *followerService) ListFollowers(userID string, limit, offset int) ([]models.User, error) {
	// TODO: Update repo call signature
	followers, err := s.followerRepo.GetFollowers(userID, limit, offset)
	if err != nil {
		log.Printf("Error listing followers in service: %v", err)
		return nil, fmt.Errorf("failed to retrieve followers")
	}
	// Optionally filter/map user data before returning
	return followers, nil
}

// ListFollowing retrieves a paginated list of users the given userID is following
func (s *followerService) ListFollowing(userID string, limit, offset int) ([]models.User, error) {
	// TODO: Update repo call signature
	following, err := s.followerRepo.GetFollowing(userID, limit, offset)
	if err != nil {
		log.Printf("Error listing following in service: %v", err)
		return nil, fmt.Errorf("failed to retrieve following list")
	}
	// Optionally filter/map user data
	return following, nil
}

// ListPendingRequests retrieves lists of pending received and sent follow requests for the given userID
func (s *followerService) ListPendingRequests(userID string) (map[string][]models.User, error) {
	// TODO: Update repo call signature/logic to get both received and sent
	received, err := s.followerRepo.GetPendingReceivedRequests(userID)
	if err != nil {
		log.Printf("Error listing pending received requests in service: %v", err)
		return nil, fmt.Errorf("failed to retrieve pending received requests")
	}

	sent, err := s.followerRepo.GetPendingSentRequests(userID)
	if err != nil {
		log.Printf("Error listing pending sent requests in service: %v", err)
		return nil, fmt.Errorf("failed to retrieve pending sent requests")
	}

	response := map[string][]models.User{
		"received": received,
		"sent":     sent,
	}

	return response, nil
}
