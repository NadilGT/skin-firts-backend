package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
)

func DB_GetFavoriteDoctors() ([]dto.DoctorInfoModel, error) {
	cursor, err := dbConfigs.DoctorInfoCollection.Find(
		context.Background(),
		map[string]interface{}{"favorite":true},
	)

	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var doctors []dto.DoctorInfoModel
	if err = cursor.All(context.Background(), &doctors); err != nil {
		return nil, err
	}

	return doctors, nil
}