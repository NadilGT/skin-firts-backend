package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
)


func DB_GetDoctorInfoByName(name string) (*dto.DoctorInfoModel, error) {
	var info dto.DoctorInfoModel

	err := dbConfigs.DoctorInfoCollection.FindOne(
		context.Background(),
		map[string]interface{}{"name": name},
	).Decode(&info)

	if err != nil {
		return nil, err
	}

	return &info, nil
}