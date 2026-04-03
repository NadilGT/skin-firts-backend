package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// DB_SearchStaff searches for staff members across admin_users, doctor_users, and staff_users collections.
func DB_SearchStaff(query dto.SearchStaffQuery) ([]dto.StaffMember, int64, error) {
	ctx := context.Background()

	// 1. Build the base filter
	filter := bson.M{}
	if query.Query != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": query.Query, "$options": "i"}},
			{"email": bson.M{"$regex": query.Query, "$options": "i"}},
			{"phoneNumber": bson.M{"$regex": query.Query, "$options": "i"}},
		}
	}

	// 2. Identify target collections based on role filter
	var mainCollection *mongo.Collection
	var otherCollections []string

	if query.Role != "" {
		switch query.Role {
		case "admin", "super_admin":
			mainCollection = dbConfigs.AdminUserCollection
		case "doctor":
			mainCollection = dbConfigs.DoctorUserCollection
		default:
			mainCollection = dbConfigs.StaffUserCollection
			// Add a direct filter for the specific role if searching staff_users
			filter["role"] = query.Role
		}
	} else {
		// Global search across all three
		mainCollection = dbConfigs.AdminUserCollection
		otherCollections = []string{"doctor_users", "staff_users"}
	}

	// 3. Build aggregation pipeline
	pipeline := mongo.Pipeline{}

	// Initial match
	pipeline = append(pipeline, bson.D{{Key: "$match", Value: filter}})

	// Union with other collections if necessary
	for _, collName := range otherCollections {
		pipeline = append(pipeline, bson.D{{Key: "$unionWith", Value: bson.M{
			"coll": collName,
			"pipeline": mongo.Pipeline{
				bson.D{{Key: "$match", Value: filter}},
			},
		}}})
	}

	// 4. Get Total Count
	countPipeline := append(pipeline, bson.D{{Key: "$count", Value: "total"}})
	countCursor, err := mainCollection.Aggregate(ctx, countPipeline)
	if err != nil {
		return nil, 0, err
	}
	
	var countResult []struct {
		Total int64 `bson:"total"`
	}
	if err := countCursor.All(ctx, &countResult); err != nil {
		return nil, 0, err
	}
	
	var total int64
	if len(countResult) > 0 {
		total = countResult[0].Total
	}

	// 5. Data query with Sort, Skip, and Limit
	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.M{"name": 1}}})
	pipeline = append(pipeline, bson.D{{Key: "$skip", Value: int64((query.Page - 1) * query.Limit)}})
	pipeline = append(pipeline, bson.D{{Key: "$limit", Value: int64(query.Limit)}})

	cursor, err := mainCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []dto.StaffMember
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}

	// Ensure we return an empty slice instead of nil for data
	if results == nil {
		results = []dto.StaffMember{}
	}

	return results, total, nil
}
