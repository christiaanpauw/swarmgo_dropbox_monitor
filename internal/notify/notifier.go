package notify

import "context"

// Notifier defines the interface for sending notifications
type Notifier interface {
	SendNotification(ctx context.Context, message string) error
}
