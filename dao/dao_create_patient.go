package dao

import (
	"context"
	"errors"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
)

// DB_CreatePatient inserts a new patient into the "patients" collection.
// Duplicate check is by email (firebaseUid is optional/legacy).
func DB_CreatePatient(patient dto.PatientUser) error {
	ctx := context.Background()

	count, err := dbConfigs.PatientCollection.CountDocuments(ctx, bson.M{"email": patient.Email})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("a patient with this email already exists")
	}

	_, err = dbConfigs.PatientCollection.InsertOne(ctx, patient)
	return err
}
