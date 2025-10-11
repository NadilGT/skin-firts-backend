package dao

import (
	"context"
	"errors"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"go.mongodb.org/mongo-driver/bson"
)

func DB_FindAllDoctors() (*[]dto.DoctorInfoModel, error) {
	var doctors []dto.DoctorInfoModel

	results, err := dbConfigs.DoctorInfoCollection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	for results.Next(context.Background()) {
		var doctor dto.DoctorInfoModel
		if err := results.Decode(&doctor); err != nil {
			return nil, errors.New("error decoding doctors")
		}
		doctors = append(doctors, doctor)
	}
	return &doctors, nil
}