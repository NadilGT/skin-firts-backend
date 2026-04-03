package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB_SearchPatients searches for patients in the patients collection.
func DB_SearchPatients(query dto.SearchPatientQuery) ([]dto.PatientUser, int64, error) {
	ctx := context.Background()

	// 1. Build the filter
	filter := bson.M{}
	
	if query.Query != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": query.Query, "$options": "i"}},
			{"email": bson.M{"$regex": query.Query, "$options": "i"}},
			{"phoneNumber": bson.M{"$regex": query.Query, "$options": "i"}},
		}
	}

	// 2. Get Total Count
	total, err := dbConfigs.PatientCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 3. Data query with Sort, Skip, and Limit
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}}) // Latest first
	findOptions.SetSkip(int64((query.Page - 1) * query.Limit))
	findOptions.SetLimit(int64(query.Limit))

	cursor, err := dbConfigs.PatientCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []dto.PatientUser
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	// Ensure empty slice instead of nil
	if results == nil {
		results = []dto.PatientUser{}
	}

	return results, total, nil
}
