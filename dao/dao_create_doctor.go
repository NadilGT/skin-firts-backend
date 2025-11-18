package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
)

func DB_CreateDoctor(doctor dto.DoctorInfoModel) error {
	_, err := dbConfigs.FeaturedLawyerCollection.InsertOne(context.Background(), doctor)
	if err != nil {
		return err
	}
	return nil
}
