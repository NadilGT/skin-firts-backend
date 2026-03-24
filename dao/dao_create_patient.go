package dao

import (
	"context"
	"errors"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
)

// DB_CreatePatient inserts a new patient into the "patients" collection.
// It returns an error if a document with the same FirebaseUID already exists.
func DB_CreatePatient(patient dto.PatientUser) error {
	ctx := context.Background()

	// Prevent duplicate Firebase UID
	count, err := dbConfigs.PatientCollection.CountDocuments(ctx, bson.M{"firebaseUid": patient.FirebaseUID})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("a patient with this Firebase UID already exists")
	}

	_, err = dbConfigs.PatientCollection.InsertOne(ctx, patient)
	return err
}
