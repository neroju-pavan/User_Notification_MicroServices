package utils

import (
	"time"

	"test123/events"

	"github.com/google/uuid"
)

func NewEmailNotificationEvent(
	userID int,
	action string, // otp | user_created | user_deleted | notify
	title string,
	message string,
	target string, // email
	metadata map[string]string,
) events.NotificationEvent {

	return events.NotificationEvent{
		EventID:          uuid.New(),
		UserID:           userID,
		Action:           action,
		NotificationType: "email",
		Title:            title,
		Message:          message,
		Target:           target,
		Metadata:         metadata,
		CreatedAt:        time.Now(),
	}
}
