package dao

import (
	"context"
	"errors"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
)

// DB_CreateDoctorUser inserts a new doctor user into the "doctor_users" collection.
// Duplicate check is by email (firebaseUid is optional/legacy).
func DB_CreateDoctorUser(doctor dto.DoctorUser) error {
	ctx := context.Background()

	count, err := dbConfigs.DoctorUserCollection.CountDocuments(ctx, bson.M{"email": doctor.Email})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("a doctor user with this email already exists")
	}

	_, err = dbConfigs.DoctorUserCollection.InsertOne(ctx, doctor)
	return err
}
