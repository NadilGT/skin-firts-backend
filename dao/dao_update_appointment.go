package dao

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"lawyerSL-Backend/dbConfigs"
)

func DB_UpdateAppointmentStatus(id primitive.ObjectID, status string) error {
	collection := dbConfigs.AppointmentCollection

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}
