package types

// WebSocketNotifier defines the interface for notification broadcasting
type WebSocketNotifier interface {
	BroadcastNotification(userID string, notificationType string)
}
