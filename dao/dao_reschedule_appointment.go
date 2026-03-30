package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson"

)

// Get appointment by ID
func DB_GetAppointmentByID(id string) (*dto.AppointmentModel, error) {
	var appointment dto.AppointmentModel
	err := dbConfigs.AppointmentCollection.FindOne(
		context.Background(),
		bson.M{"appointmentId": id},
	).Decode(&appointment)

	if err != nil {
		return nil, err
	}

	return &appointment, nil
}

// Check availability excluding a specific appointment (for rescheduling)
func DB_RescheduleAppointment(id string, newDate time.Time) error {
	update := bson.M{
		"$set": bson.M{
			"appointmentDate": newDate,
			"updatedAt":       time.Now(),
		},
	}

	_, err := dbConfigs.AppointmentCollection.UpdateOne(
		context.Background(),
		bson.M{"appointmentId": id},
		update,
	)

	return err
}