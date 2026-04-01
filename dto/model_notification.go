package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationModel represents a push notification stored in MongoDB.
// Cursor-based pagination is supported via the `_id` field (ObjectID).
type NotificationModel struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	NotificationID string             `json:"notificationId" bson:"notificationId"`
	UserID         string             `json:"userId" bson:"userId"`
	Title          string             `json:"title" bson:"title"`
	Body           string             `json:"body" bson:"body"`
	Type           string             `json:"type" bson:"type"`
	Data           map[string]string  `json:"data,omitempty" bson:"data,omitempty"`
	IsRead         bool               `json:"isRead" bson:"isRead"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt"`
}
