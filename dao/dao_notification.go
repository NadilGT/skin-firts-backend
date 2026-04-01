package dao

import (
	"context"
	"fmt"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB_SaveNotification inserts a new notification into the notifications collection.
// The notificationId is generated via GenerateId before calling this function.
func DB_SaveNotification(notification dto.NotificationModel) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := dbConfigs.NotificationCollection.InsertOne(ctx, notification)
	if err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}
	return nil
}

// DB_GetNotificationsByUserID returns up to `limit` notifications for a user,
// using cursor-based (WhatsApp-style) pagination.
//
//   - If lastId is empty → returns the latest `limit` notifications.
//   - If lastId is provided → returns notifications OLDER than that ObjectID.
//
// Results are always sorted newest-first (_id descending).
func DB_GetNotificationsByUserID(userID string, lastID string, limit int64) ([]dto.NotificationModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}

	// Cursor-based pagination: fetch records with _id less than lastID (older)
	if lastID != "" {
		oid, err := primitive.ObjectIDFromHex(lastID)
		if err != nil {
			return nil, fmt.Errorf("invalid lastId format: %w", err)
		}
		filter["_id"] = bson.M{"$lt": oid}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}). // newest first
		SetLimit(limit)

	cursor, err := dbConfigs.NotificationCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer cursor.Close(ctx)

	var notifications []dto.NotificationModel
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, fmt.Errorf("failed to decode notifications: %w", err)
	}

	return notifications, nil
}

// DB_MarkNotificationRead sets isRead = true for a given notificationId.
func DB_MarkNotificationRead(notificationID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"notificationId": notificationID}
	update := bson.M{"$set": bson.M{"isRead": true}}

	result, err := dbConfigs.NotificationCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("no notification found with id: %s", notificationID)
	}
	return nil
}

// DB_MarkAllNotificationsRead sets isRead = true for all notifications of a user.
func DB_MarkAllNotificationsRead(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID, "isRead": false}
	update := bson.M{"$set": bson.M{"isRead": true}}

	_, err := dbConfigs.NotificationCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}
	return nil
}
