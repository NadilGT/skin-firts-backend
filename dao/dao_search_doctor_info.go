package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB_SearchDoctorInfo searches for doctors in the doctor_info collection.
func DB_SearchDoctorInfo(query dto.SearchDoctorInfoQuery, branchId string) ([]dto.DoctorInfoModel, int64, error) {
	ctx := context.Background()

	// 1. Build the filter
	filter := bson.M{}
	
	if query.Query != "" {
		filter["name"] = bson.M{"$regex": query.Query, "$options": "i"}
	}
	
	if query.FocusId != "" {
		filter["focus_id"] = query.FocusId
	} else if query.Focus != "" {
		filter["focus"] = bson.M{"$regex": query.Focus, "$options": "i"}
	}
	
	if query.Special != "" {
		filter["special"] = bson.M{"$regex": query.Special, "$options": "i"}
	}
	
	if branchId != "" {
		filter["branchIds"] = branchId
	}

	// 2. Get Total Count
	total, err := dbConfigs.DoctorInfoCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 3. Data query with Sort, Skip, and Limit
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "name", Value: 1}}) // Alphabetical order
	findOptions.SetSkip(int64((query.Page - 1) * query.Limit))
	findOptions.SetLimit(int64(query.Limit))

	cursor, err := dbConfigs.DoctorInfoCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []dto.DoctorInfoModel
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	// Ensure empty slice instead of nil
	if results == nil {
		results = []dto.DoctorInfoModel{}
	}

	return results, total, nil
}
