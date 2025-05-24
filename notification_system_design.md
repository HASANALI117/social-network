# Notification System Backend Design

This document outlines the backend core design for a new real-time notification system.

## 1. Notification Data Model (`Notification` struct)

A `Notification` struct will be defined in `backend/pkg/models/notification.go`.

```go
package models

import (
	"time"
)

// NotificationType defines the type of notification.
type NotificationType string

// EntityType defines the type of entity a notification might refer to.
type EntityType string

const (
	FollowRequestNotification NotificationType = "follow_request"
	GroupInviteNotification   NotificationType = "group_invite"
	GroupJoinRequestNotification NotificationType = "group_join_request"
	GroupEventCreatedNotification NotificationType = "group_event_created"
	// Add other notification types here in the future
)

const (
	UserEntityType  EntityType = "user"
	GroupEntityType EntityType = "group"
	EventEntityType EntityType = "event"
	// Add other entity types here
)

// Notification represents a notification in the system.
type Notification struct {
	ID         string           `json:"id" db:"id"`
	UserID     string           `json:"user_id" db:"user_id"`           // Recipient of the notification
	Type       NotificationType `json:"type" db:"type"`                 // Type of notification (e.g., "follow_request")
	EntityType EntityType       `json:"entity_type" db:"entity_type"`   // Type of the entity this notification refers to
	Message    string           `json:"message" db:"message"`             // User-friendly message
	EntityID   string           `json:"entity_id" db:"entity_id"`         // ID of the related entity (e.g., follower's UserID, GroupID, EventID)
	IsRead     bool             `json:"is_read" db:"is_read"`             // Whether the notification has been read
	CreatedAt  time.Time        `json:"created_at" db:"created_at"`       // Timestamp of creation
}
```

**Key considerations:**
*   `ID`: Will be a UUID string.
*   `UserID`: String, referencing `User.ID`.
*   `Type`: `NotificationType` (string-based for extensibility) to categorize notifications.
*   `EntityType`: `EntityType` (string-based) to specify what kind of entity `EntityID` refers to.
*   `Message`: A human-readable string.
*   `EntityID`: String, the ID of the relevant entity (e.g., the user who sent a follow request, the group an invite is for).
*   `IsRead`: Boolean, defaults to `false`.
*   `CreatedAt`: `time.Time`.

## 2. Database Schema (`notifications` table)

The SQL schema for the `notifications` table:

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,                     -- Recipient User ID
    type VARCHAR(255) NOT NULL,
    entity_type VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    entity_id UUID NOT NULL,                   -- ID of the related entity
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    -- Optional: Add foreign keys for entity_id if they always refer to a specific table
    -- based on entity_type, though this is harder to enforce directly in SQL
    -- for a polymorphic association. Application-level integrity would be key.
);

-- Indexes for performance
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_user_id_is_read ON notifications(user_id, is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
```

## 3. Repository Layer (`NotificationRepository` interface)

This interface will be defined in `backend/pkg/repositories/notification_repository.go`.

```go
package repositories

import (
	"context"
	"github.com/HASANALI117/social-network/pkg/models" // Adjust import path as needed
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *models.Notification) error
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error)
	MarkAsRead(ctx context.Context, notificationID string, userID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}
```

## 4. Service Layer (`NotificationService` interface)

This interface will be defined in `backend/pkg/services/notification_service.go`.

```go
package services

import (
	"context"
	"github.com/HASANALI117/social-network/pkg/models" // Adjust import path as needed
)

type NotificationService interface {
	CreateNotification(ctx context.Context, recipientID string, nType models.NotificationType, eType models.EntityType, message string, entityID string) (*models.Notification, error)
	GetUserNotifications(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error)
	MarkNotificationAsRead(ctx context.Context, notificationID string, userID string) error
	MarkAllUserNotificationsAsRead(ctx context.Context, userID string) error
	GetUnreadNotificationCount(ctx context.Context, userID string) (int, error)

    // Method to assist with real-time delivery
	SendNotificationToUser(userID string, notification *models.Notification) error
}
```

## 5. Integration Points for Notification Creation

Calls to `NotificationService.CreateNotification` will be integrated into existing services:

*   **Follow Request** (e.g., in `UserService`):
    *   When `UserA` requests to follow `UserB` (who has a private profile).
    *   `recipientID`: `UserB.ID`
    *   `nType`: `models.FollowRequestNotification`
    *   `eType`: `models.UserEntityType`
    *   `message`: `"{UserA.Username} wants to follow you."`
    *   `entityID`: `UserA.ID`

*   **Group Invitation** (e.g., in `GroupService`):
    *   When `UserA` invites `UserB` to `GroupX`.
    *   `recipientID`: `UserB.ID`
    *   `nType`: `models.GroupInviteNotification`
    *   `eType`: `models.GroupEntityType`
    *   `message`: `"{UserA.Username} invited you to join {GroupX.Name}."`
    *   `entityID`: `GroupX.ID`

*   **Group Join Request** (e.g., in `GroupService`):
    *   When `UserA` requests to join `GroupX` (created by `UserB`).
    *   `recipientID`: `UserB.ID` (Group Creator)
    *   `nType`: `models.GroupJoinRequestNotification`
    *   `eType`: `models.UserEntityType`
    *   `message`: `"{UserA.Username} wants to join your group {GroupX.Name}."`
    *   `entityID`: `UserA.ID` (the requester)

*   **New Group Event** (e.g., in `GroupService`):
    *   When an `EventY` is created in `GroupX`.
    *   This will iterate over all members of `GroupX`. For each `MemberZ`:
        *   `recipientID`: `MemberZ.ID`
        *   `nType`: `models.GroupEventCreatedNotification`
        *   `eType`: `models.EventEntityType`
        *   `message`: `"A new event '{EventY.Name}' has been created in {GroupX.Name}."`
        *   `entityID`: `EventY.ID`

## 6. Real-time Delivery Plan

Modifications to `backend/pkg/websocket/hub.go`:

1.  **Hub Dependencies:**
    *   `NotificationService` will need a way to send messages, possibly through a reference to the `Hub` or an interface the `Hub` implements.

2.  **Sending Mechanism:**
    *   After a notification is created by `NotificationService.CreateNotification`, the service will call its `SendNotificationToUser` method.
    *   This method (or the Hub directly) will look up the `client` by `userID`.
    *   If the client is connected, the notification payload is sent via `client.Send`.

3.  **WebSocket Message Format:**
    ```json
    {
      "type": "new_notification",
      "payload": {
        "id": "uuid-string-for-notification",
        "user_id": "uuid-string-for-recipient",
        "type": "follow_request",
        "entity_type": "user",
        "message": "UserX wants to follow you.",
        "entity_id": "uuid-string-for-UserX",
        "is_read": false,
        "created_at": "2023-10-27T10:00:00Z"
      }
    }
    ```

4.  **Hub Modifications (`hub.go`):**
    *   The `Client` struct's `Send` channel (`chan interface{}`) is suitable.

## 7. Diagrams

### Class Diagram (Conceptual)

```mermaid
classDiagram
    class Notification {
        +string ID
        +string UserID
        +NotificationType Type
        +EntityType EntityType
        +string Message
        +string EntityID
        +bool IsRead
        +time.Time CreatedAt
    }

    class NotificationRepository {
        <<Interface>>
        +Create(ctx, notification) error
        +GetByUserID(ctx, userID, limit, offset) []*Notification, error
        +MarkAsRead(ctx, notificationID, userID) error
        +MarkAllAsRead(ctx, userID) error
        +GetUnreadCount(ctx, userID) int, error
    }

    class NotificationService {
        <<Interface>>
        +CreateNotification(ctx, recipientID, type, entityType, message, entityID) *Notification, error
        +GetUserNotifications(ctx, userID, limit, offset) []*Notification, error
        +MarkNotificationAsRead(ctx, notificationID, userID) error
        +MarkAllUserNotificationsAsRead(ctx, userID) error
        +GetUnreadNotificationCount(ctx, userID) int, error
        +SendNotificationToUser(userID, notification) error
    }

    class Hub {
        +Clients map[string]*Client
        +Register chan *Client
        +Unregister chan *Client
        +Broadcast chan *Message // Existing chat message
        +SendToUser(userID string, payload interface{}) // New or adapted method
    }

    class Client {
        +string UserID
        +Hub *Hub
        +Conn *websocket.Conn
        +Send chan interface{} // Can send various payload types
    }

    NotificationService ..> NotificationRepository : uses
    NotificationService ..> Hub : uses (for real-time push)
    Hub o-- Client : manages
```

### Sequence Diagram: New Follow Request Notification

```mermaid
sequenceDiagram
    participant UserA_Client
    participant Backend_UserService
    participant Backend_NotificationService
    participant Backend_NotificationRepository
    participant Database
    participant Backend_WebSocketHub
    participant UserB_Client

    UserA_Client->>+Backend_UserService: HTTP POST /users/{UserB_ID}/follow
    Backend_UserService->>+Backend_NotificationService: CreateNotification(recipientID=UserB_ID, type="follow_request", entityType="user", message="UserA wants to follow you", entityID=UserA_ID)
    Backend_NotificationService->>+Backend_NotificationRepository: Create(notification)
    Backend_NotificationRepository->>+Database: INSERT into notifications
    Database-->>-Backend_NotificationRepository: Success
    Backend_NotificationRepository-->>-Backend_NotificationService: Success, returns Notification
    Backend_NotificationService-->> Backend_UserService: Returns Notification
    Backend_UserService-->>-UserA_Client: HTTP 200 OK

    alt UserB is connected via WebSocket
        Backend_NotificationService->>+Backend_WebSocketHub: SendToUser(UserB_ID, new_notification_payload)
        Backend_WebSocketHub->>UserB_Client: Sends WebSocket Message (new_notification)
        UserB_Client-->>-Backend_WebSocketHub: (Receives message)
    end