package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
)

func DB_GetAppointmentByAppointmentID(appointmentID string, branchId string) (dto.AppointmentModel, error) {
	var appointment dto.AppointmentModel

	filter := bson.M{"appointmentId": appointmentID}
	if branchId != "" {
		filter["branchId"] = branchId
	}

	err := dbConfigs.AppointmentCollection.FindOne(context.Background(), filter).Decode(&appointment)
	if err != nil {
		return appointment, err
	}

	return appointment, nil
}
