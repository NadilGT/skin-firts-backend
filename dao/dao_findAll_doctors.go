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

func DB_FindDoctorsByFocus(focusId string, branchId string) (*[]dto.DoctorInfoModel, error) {
	var doctors []dto.DoctorInfoModel

	filter := bson.M{"focus_id": focusId}
	if branchId != "" {
		filter["branchIds"] = branchId
	}
	results, err := dbConfigs.DoctorInfoCollection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer results.Close(context.Background())

	for results.Next(context.Background()) {
		var doctor dto.DoctorInfoModel
		if err := results.Decode(&doctor); err != nil {
			return nil, errors.New("error decoding doctors")
		}
		doctors = append(doctors, doctor)
	}
	return &doctors, nil
}