package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
)

func DB_CreateDoctorInfo(info dto.DoctorInfoModel) error {
	_, err := dbConfigs.DoctorInfoCollection.InsertOne(context.Background(), info)
	if err != nil {
		return err
	}
	return nil
}
