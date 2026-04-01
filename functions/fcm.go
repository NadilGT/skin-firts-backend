package functions

import (
	"context"
	"fmt"
	"log"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// SaveAndSendNotification is the single internal helper the backend calls
// whenever it needs to notify a user.
//
// Flow: MongoDB save (guaranteed) → FCM send (best-effort, logged on failure)
//
// Parameters:
//   - firebaseApp : the initialised Firebase app
//   - fcmToken    : device token (empty = skip FCM, still saves to DB)
//   - userID      : Firebase UID of the recipient
//   - title       : notification title
//   - body        : notification body
//   - notifType   : e.g. "REPORT_READY", "APPOINTMENT_CONFIRMED"
//   - data        : optional key-value payload for the Flutter app
func SaveAndSendNotification(
	firebaseApp *firebase.App,
	fcmToken string,
	userID string,
	title string,
	body string,
	notifType string,
	data map[string]string,
) error {
	ctx := context.Background()

	// 1️⃣  Generate a human-readable ID (e.g. NOTIF-001)
	notificationId, err := dao.GenerateId(ctx, "notifications", "NOTIF")
	if err != nil {
		notificationId = fmt.Sprintf("NOTIF-%d", time.Now().UnixNano())
		log.Printf("⚠️  Notification ID counter failed, using fallback: %s", notificationId)
	}

	// 2️⃣  Persist to MongoDB FIRST — data is never lost even if FCM fails
	notification := dto.NotificationModel{
		ID:             primitive.NewObjectID(),
		NotificationID: notificationId,
		UserID:         userID,
		Title:          title,
		Body:           body,
		Type:           notifType,
		Data:           data,
		IsRead:         false,
		CreatedAt:      time.Now(),
	}

	if err := dao.DB_SaveNotification(notification); err != nil {
		return fmt.Errorf("failed to persist notification to DB: %w", err)
	}

	log.Printf("✅ Notification saved to DB (id=%s, user=%s)", notificationId, userID)

	// 3️⃣  Send FCM push (best-effort — failure is logged, not fatal)
	if fcmToken != "" {
		if err := SendFCMNotification(firebaseApp, fcmToken, title, body, data); err != nil {
			log.Printf("⚠️  FCM send failed for notification %s: %v", notificationId, err)
			// We don't return an error here — the notification is already in DB
		}
	} else {
		log.Printf("ℹ️  FCM skipped for notification %s: no device token for user %s", notificationId, userID)
	}

	return nil
}

