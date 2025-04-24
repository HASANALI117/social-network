package services

import (
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
	ListFollowers(userID string) ([]models.User, error)
	ListFollowing(userID string) ([]models.User, error)
	ListPendingRequests(userID string) ([]models.User, error)
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

	// Check if target user exists (optional, depends on requirements)
	// _, err := s.userRepo.FindByID(targetID)
	// if err != nil {
	//     if err == sql.ErrNoRows {
	//         return errors.New("target user not found")
	//     }
	//     log.Printf("Error checking target user: %v", err)
	//     return fmt.Errorf("internal server error")
	// }

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

	// Create the follow request
	err = s.followerRepo.CreateFollowRequest(requesterID, targetID)
	if err != nil {
		log.Printf("Error creating follow request in service: %v", err)
		return fmt.Errorf("failed to send follow request")
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

// ListFollowers retrieves the list of users following the given userID
func (s *followerService) ListFollowers(userID string) ([]models.User, error) {
	followers, err := s.followerRepo.GetFollowers(userID)
	if err != nil {
		log.Printf("Error listing followers in service: %v", err)
		return nil, fmt.Errorf("failed to retrieve followers")
	}
	// Optionally filter/map user data before returning
	return followers, nil
}

// ListFollowing retrieves the list of users the given userID is following
func (s *followerService) ListFollowing(userID string) ([]models.User, error) {
	following, err := s.followerRepo.GetFollowing(userID)
	if err != nil {
		log.Printf("Error listing following in service: %v", err)
		return nil, fmt.Errorf("failed to retrieve following list")
	}
	// Optionally filter/map user data
	return following, nil
}

// ListPendingRequests retrieves the list of pending follow requests for the given userID
func (s *followerService) ListPendingRequests(userID string) ([]models.User, error) {
	requests, err := s.followerRepo.GetPendingRequests(userID)
	if err != nil {
		log.Printf("Error listing pending requests in service: %v", err)
		return nil, fmt.Errorf("failed to retrieve pending requests")
	}
	// Optionally filter/map user data
	return requests, nil
}
