package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DB_FindAllAppointments() ([]dto.AppointmentModel, error) {
	var appointments []dto.AppointmentModel

	findOptions := options.Find().SetSort(bson.D{{"createdAt", -1}})

	cursor, err := dbConfigs.AppointmentCollection.Find(context.Background(), bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var appointment dto.AppointmentModel
		if err := cursor.Decode(&appointment); err != nil {
			return nil, err
		}
		appointments = append(appointments, appointment)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return appointments, nil
}
