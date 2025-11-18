package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func DB_CreateAppointment(appointment dto.AppointmentModel) error {
	appointment.CreatedAt = time.Now()
	appointment.UpdatedAt = time.Now()

	_, err := dbConfigs.AppointmentCollection.InsertOne(context.Background(), appointment)
	return err
}

func DB_IsTimeSlotAvailable(doctorID string, date time.Time, timeSlot string) (bool, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	count, err := dbConfigs.AppointmentCollection.CountDocuments(
		context.Background(),
		bson.M{
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
