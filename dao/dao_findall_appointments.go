package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB_FindAllAppointments returns a paginated list of all appointments sorted by createdAt desc.
// page is 1-indexed. Returns the slice, total document count, and any error.
func DB_FindAllAppointments(page, limit int) ([]dto.AppointmentModel, int64, error) {
	ctx := context.Background()
	filter := bson.M{}

	total, err := dbConfigs.AppointmentCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	findOptions := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(limit))

	cursor, err := dbConfigs.AppointmentCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var appointments []dto.AppointmentModel
	for cursor.Next(ctx) {
		var a dto.AppointmentModel
		if err := cursor.Decode(&a); err != nil {
			return nil, 0, err
		}
		appointments = append(appointments, a)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	return appointments, total, nil
}

// DB_FindAppointmentsByDoctorID returns a paginated list of appointments for a specific doctor.
func DB_FindAppointmentsByDoctorID(doctorID string, page, limit int) ([]dto.AppointmentModel, int64, error) {
	ctx := context.Background()
	filter := bson.M{"doctorId": doctorID}

	total, err := dbConfigs.AppointmentCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	findOptions := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(limit))

	cursor, err := dbConfigs.AppointmentCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var appointments []dto.AppointmentModel
	for cursor.Next(ctx) {
		var a dto.AppointmentModel
		if err := cursor.Decode(&a); err != nil {
			return nil, 0, err
		}
		appointments = append(appointments, a)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	return appointments, total, nil
}
// DB_FindAppointmentsByPatientID returns a paginated list of appointments for a specific patient (Firebase ID).
func DB_FindAppointmentsByPatientID(patientID string, page, limit int) ([]dto.AppointmentModel, int64, error) {
	ctx := context.Background()
	filter := bson.M{"patientId": patientID}

	total, err := dbConfigs.AppointmentCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * limit)
	findOptions := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(limit))

	cursor, err := dbConfigs.AppointmentCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var appointments []dto.AppointmentModel
	for cursor.Next(ctx) {
		var a dto.AppointmentModel
		if err := cursor.Decode(&a); err != nil {
			return nil, 0, err
		}
		appointments = append(appointments, a)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	return appointments, total, nil
}
