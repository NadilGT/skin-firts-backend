package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
)

func DB_UpdateDoctorInfoByDoctorId(doctorID string, info dto.DoctorInfoModel) error {
	filter := bson.M{"doctor_id": doctorID}
	update := bson.M{
		"$set": bson.M{
			"name":        info.Name,
			"experience":  info.Experience,
			"focus":       info.Focus,
			"special":     info.Special,
			"starts":      info.Starts,
			"messages":    info.Messages,
			"date":        info.Date,
			"profile":     info.Profile,
			"career":      info.Career,
			"highlights":  info.Highlights,
			"favorite":    info.Favorite,
			"profile_pic": info.ProfilePic,
			"branchIds":   info.BranchIds,
			"status":      info.Status,
			"updatedAt":   info.UpdatedAt,
		},
	}

	_, err := dbConfigs.DoctorInfoCollection.UpdateOne(context.Background(), filter, update)
	return err
}

func DB_AddDoctorToBranch(doctorID string, branchID string) error {
	filter := bson.M{"doctor_id": doctorID}
	update := bson.M{
		"$addToSet": bson.M{
			"branchIds": branchID,
		},
	}
	_, err := dbConfigs.DoctorInfoCollection.UpdateOne(context.Background(), filter, update)
	return err
}

func DB_RemoveDoctorFromBranch(doctorID string, branchID string) error {
	filter := bson.M{"doctor_id": doctorID}
	update := bson.M{
		"$pull": bson.M{
			"branchIds": branchID,
		},
	}
	_, err := dbConfigs.DoctorInfoCollection.UpdateOne(context.Background(), filter, update)
	return err
}
