package functions

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
)

// SendFCMNotification sends a push notification to a single device token.
// It uses the already-initialised Firebase app passed in from main.
// If the token is empty the function logs a warning and returns nil (not an error).
func SendFCMNotification(firebaseApp *firebase.App, token string, title string, body string, data map[string]string) error {
	if token == "" {
		log.Println("⚠️  FCM: device token is empty – skipping notification")
		return nil
	}

	ctx := context.Background()
	client, err := firebaseApp.Messaging(ctx)
	if err != nil {
		return fmt.Errorf("FCM: failed to get messaging client: %w", err)
	}

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data, // optional key-value payload for the Flutter app
	}

	msgID, err := client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("FCM: send failed: %w", err)
	}

	log.Printf("✅ FCM: notification sent successfully (id=%s)", msgID)
	return nil
}
