package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
)

func DB_ToggleFavoriteDoctor(name string) (*dto.DoctorInfoModel, error) {
	var updatedDoctor dto.DoctorInfoModel

	filter := map[string]interface{}{"name": name}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"favorite": bson.M{
				"$not": "$favorite",
			},
		},
	}

	var current dto.DoctorInfoModel
	err := dbConfigs.DoctorInfoCollection.FindOne(context.Background(), filter).Decode(&current)
	if err != nil {
		return nil, err
	}

	newFavorite := !current.Favorite
	update = map[string]interface{}{
		"$set": map[string]interface{}{"favorite": newFavorite},
	}

	err = dbConfigs.DoctorInfoCollection.FindOneAndUpdate(
		context.Background(),
		filter,
		update,
	).Decode(&updatedDoctor)

	if err != nil {
		return nil, err
	}

	updatedDoctor.Favorite = newFavorite
	return &updatedDoctor, nil
}
