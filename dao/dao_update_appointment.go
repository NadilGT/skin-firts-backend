package dao

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"lawyerSL-Backend/dbConfigs"
)

func DB_UpdateAppointmentStatus(id string, status string, branchId string) error {
	collection := dbConfigs.AppointmentCollection

	filter := bson.M{"appointmentId": id}
	if branchId != "" {
		filter["branchId"] = branchId
	}
	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}

func DB_GetRunningAppointment(doctorID string, date time.Time, branchId string) (int, error) {
	collection := dbConfigs.AppointmentCollection

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	filter := bson.M{
		"doctorId": doctorID,
		"branchId": branchId,
		"status":   "running",
		"appointmentDate": bson.M{
			"$gte": startOfDay,
			"$lt":  endOfDay,
		},
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
