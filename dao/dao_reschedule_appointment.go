package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Get appointment by ID
func DB_GetAppointmentByID(id string) (*dto.AppointmentModel, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var appointment dto.AppointmentModel
	err = dbConfigs.AppointmentCollection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&appointment)

	if err != nil {
		return nil, err
	}

	return &appointment, nil
}

// Check availability excluding a specific appointment (for rescheduling)
func DB_IsTimeSlotAvailableExcluding(doctorID string, date time.Time, timeSlot string, excludeAppointmentID string) (bool, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	excludeID, err := primitive.ObjectIDFromHex(excludeAppointmentID)
	if err != nil {
		return false, err
	}

	count, err := dbConfigs.AppointmentCollection.CountDocuments(
		context.Background(),
		bson.M{
			"_id": bson.M{"$ne": excludeID}, // Exclude the current appointment
			"doctorId": doctorID,
			"appointmentDate": bson.M{
				"$gte": startOfDay,
				"$lt":  endOfDay,
			},
			"timeSlot": timeSlot,
			"status": bson.M{
				"$nin": []string{"cancelled"},
			},
		},
	)

	if err != nil {
		return false, err
	}

	return count == 0, nil
}

// Reschedule appointment
func DB_RescheduleAppointment(id string, newDate time.Time) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"appointmentDate": newDate,
			"updatedAt":       time.Now(),
		},
	}

	_, err = dbConfigs.AppointmentCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		update,
	)

	return err
}