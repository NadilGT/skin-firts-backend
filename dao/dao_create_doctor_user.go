package dao

import (
	"context"
	"errors"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
)

// DB_CreateDoctorUser inserts a new doctor user into the "doctor_users" collection.
// It returns an error if a document with the same FirebaseUID already exists.
func DB_CreateDoctorUser(doctor dto.DoctorUser) error {
	ctx := context.Background()

	// Prevent duplicate Firebase UID
	count, err := dbConfigs.DoctorUserCollection.CountDocuments(ctx, bson.M{"firebaseUid": doctor.FirebaseUID})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("a doctor user with this Firebase UID already exists")
	}

	_, err = dbConfigs.DoctorUserCollection.InsertOne(ctx, doctor)
	return err
}
