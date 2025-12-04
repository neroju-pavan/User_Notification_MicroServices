package events

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type NotificationEvent struct {
	EventID          uuid.UUID         `json:"event_id"`
	UserID           int               `json:"user_id"`
	NotificationType string            `json:"notification_type"` // email | sms | push
	Action           string            `json:"action"`
	Title            string            `json:"title,omitempty"`
	Message          string            `json:"message"`
	Target           string            `json:"target"` // email / phone / deviceToken
	Metadata         map[string]string `json:"metadata,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
}

func (e NotificationEvent) Validate() error {
	if e.UserID < 0 {
		return fmt.Errorf("user_id required")
	}
	if e.NotificationType == "" {
		return fmt.Errorf("notification_type required")
	}
	if e.Message == "" {
		return fmt.Errorf("message required")
	}
	if e.Target == "" {
		return fmt.Errorf("target required")
	}
	if e.Action == "" {
		return fmt.Errorf("Action Required")
	}
	return nil
}
