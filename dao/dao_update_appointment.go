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

func DB_GetRunningAppointment(doctorID string) (int, error) {
	collection := dbConfigs.AppointmentCollection

	filter := bson.M{
		"doctorId": doctorID,
		"status":   "running",
	}

	var result map[string]interface{}
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return 0, err
	}

	if num, ok := result["appointmentNumber"].(int32); ok {
		return int(num), nil
	} else if num, ok := result["appointmentNumber"].(float64); ok {
		return int(num), nil
	} else if num, ok := result["appointmentNumber"].(int); ok {
		return num, nil
	}

	return 0, nil
}
