package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DB_CreateBranch(branch dto.BranchModel) error {
	_, err := dbConfigs.BranchCollection.InsertOne(context.Background(), branch)
	return err
}

func DB_GetAllBranches() ([]dto.BranchModel, error) {
	ctx := context.Background()
	cursor, err := dbConfigs.BranchCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var branches []dto.BranchModel
	if err = cursor.All(ctx, &branches); err != nil {
		return nil, err
	}
	return branches, nil
}

func DB_GetBranchByBranchId(branchId string) (*dto.BranchModel, error) {
	var branch dto.BranchModel
	err := dbConfigs.BranchCollection.FindOne(context.Background(), bson.M{"branchId": branchId}).Decode(&branch)
	if err != nil {
		return nil, err
	}
	return &branch, nil
}

func DB_UpdateBranch(branchId string, branch dto.BranchModel) error {
	filter := bson.M{"branchId": branchId}
	update := bson.M{
		"$set": bson.M{
			"name":         branch.Name,
			"address":      branch.Address,
			"phone":        branch.Phone,
			"email":        branch.Email,
			"isMainBranch": branch.IsMainBranch,
			"status":       branch.Status,
			"updatedAt":    time.Now(),
		},
	}
	_, err := dbConfigs.BranchCollection.UpdateOne(context.Background(), filter, update)
	return err
}

func DB_DeleteBranch(branchId string) error {
	_, err := dbConfigs.BranchCollection.DeleteOne(context.Background(), bson.M{"branchId": branchId})
	return err
}


func DB_SearchBranches(status string) ([]dto.BranchModel, error) {
	ctx := context.Background()
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}
	findOpts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cursor, err := dbConfigs.BranchCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var branches []dto.BranchModel
	if err = cursor.All(ctx, &branches); err != nil {
		return nil, err
	}
	return branches, nil
}
