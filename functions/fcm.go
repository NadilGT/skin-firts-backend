package functions

import (
	"context"
	"fmt"
	"log"
	"time"

	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SendFCMNotification is a no-op in local/offline mode.
// FCM requires an active internet connection and Firebase project.
// Notifications are still SAVED to MongoDB (notifications collection) for
// in-app polling by the Flutter client.
func SendFCMNotification(firebaseApp interface{}, token string, title string, body string, data map[string]string) error {
	if token == "" {
		log.Println("ℹ️  FCM: device token is empty — notification saved to DB only (offline mode)")
		return nil
	}
	// In offline mode: just log and skip
	log.Printf("ℹ️  FCM: offline mode — skipping push to token [%s...]. Notification saved to DB.", truncateToken(token))
	return nil
}

// SaveAndSendNotification persists the notification to MongoDB then attempts
// an FCM push (best-effort, silently skipped in offline mode).
//
// Parameters:
//   - firebaseApp : ignored in offline mode (kept for call-site compatibility)
//   - fcmToken    : device token (empty or ignored in offline mode)
//   - userID      : app-level user ID of the recipient
//   - title       : notification title
//   - body        : notification body text
//   - notifType   : e.g. "REPORT_READY", "APPOINTMENT_CONFIRMED"
//   - data        : optional key-value payload for the Flutter app
func SaveAndSendNotification(
	firebaseApp interface{},
	fcmToken string,
	userID string,
	title string,
	body string,
	notifType string,
	data map[string]string,
) error {
	ctx := context.Background()

	// 1. Generate a human-readable notification ID
	notificationId, err := dao.GenerateId(ctx, "notifications", "NOTIF")
	if err != nil {
		notificationId = fmt.Sprintf("NOTIF-%d", time.Now().UnixNano())
		log.Printf("⚠️  Notification ID counter failed, using fallback: %s", notificationId)
	}

	// 2. Persist to MongoDB FIRST — never lost even if FCM fails or is offline
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

	// 3. FCM push — silently skipped in offline mode
	_ = SendFCMNotification(firebaseApp, fcmToken, title, body, data)

	return nil
}

func truncateToken(token string) string {
	if len(token) > 10 {
		return token[:10]
	}
	return token
}
