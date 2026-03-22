package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DB_CreateFocus(focus *dto.FocusModel) error {
	focus.ID = primitive.NewObjectID()
	_, err := dbConfigs.FocusCollection.InsertOne(context.Background(), focus)
	return err
}

func DB_GetAllFocuses() ([]dto.FocusModel, error) {
	var focuses []dto.FocusModel

	cursor, err := dbConfigs.FocusCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &focuses); err != nil {
		return nil, err
	}

	return focuses, nil
}
