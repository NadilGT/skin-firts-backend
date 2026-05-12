package dao

import (
	"context"
	"time"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB_CreateHospitalBill inserts a new hospital bill record into the hospital_bills collection.
func DB_CreateHospitalBill(bill *dto.HospitalBillModel) error {
	if bill.ID == primitive.NilObjectID {
		bill.ID = primitive.NewObjectID()
	}
	_, err := dbConfigs.HospitalBillCollection.InsertOne(context.Background(), bill)
	return err
}

// DB_ConfirmHospitalBill updates a hospital bill's confirm status to true.
func DB_ConfirmHospitalBill(hospitalBillId string, branchId string) error {
	filter := bson.M{"hospitalBillId": hospitalBillId}
	if branchId != "" {
		filter["branchId"] = branchId
	}
	update := bson.M{"$set": bson.M{"confirm": true}}
	_, err := dbConfigs.HospitalBillCollection.UpdateOne(context.Background(), filter, update)
	return err
}

// DB_GetHospitalBill gets a hospital bill by its bill ID.
func DB_GetHospitalBill(hospitalBillId string, branchId string) (*dto.HospitalBillModel, error) {
	var bill dto.HospitalBillModel
	filter := bson.M{"hospitalBillId": hospitalBillId}
	if branchId != "" {
		filter["branchId"] = branchId
	}
	err := dbConfigs.HospitalBillCollection.FindOne(context.Background(), filter).Decode(&bill)
	if err != nil {
		return nil, err
	}
	return &bill, nil
}

// DB_SearchHospitalBillsReport searches hospital bills and returns the list, total count, and total sum of amounts.
func DB_SearchHospitalBillsReport(query dto.SearchHospitalBillQuery) ([]dto.HospitalBillModel, int64, float64, error) {
	ctx := context.Background()
	filter := bson.M{}

	if query.BranchId != "" {
		filter["branchId"] = query.BranchId
	}
	if query.DoctorId != "" {
		filter["doctorId"] = query.DoctorId
	}

	// Date range on createdAt
	if query.From != "" || query.To != "" {
		dateFilter := bson.M{}
		if query.From != "" {
			if t, err := time.Parse("2006-01-02", query.From); err == nil {
				dateFilter["$gte"] = t
			}
		}
		if query.To != "" {
			if t, err := time.Parse("2006-01-02", query.To); err == nil {
				dateFilter["$lte"] = t.Add(24 * time.Hour)
			}
		}
		if len(dateFilter) > 0 {
			filter["createdAt"] = dateFilter
		}
	}

	totalCount, err := dbConfigs.HospitalBillCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, 0, err
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}

	findOpts := options.Find().
		SetSkip(int64((query.Page - 1) * query.Limit)).
		SetLimit(int64(query.Limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := dbConfigs.HospitalBillCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, 0, err
	}
	defer cursor.Close(ctx)

	var bills []dto.HospitalBillModel
	if err = cursor.All(ctx, &bills); err != nil {
		return nil, 0, 0, err
	}

	// Sum the totalAmount for all matching records
	var totalSum float64
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":         nil,
			"totalAmount": bson.M{"$sum": "$totalAmount"},
		}}},
	}
	aggCursor, err := dbConfigs.HospitalBillCollection.Aggregate(ctx, pipeline)
	if err == nil {
		defer aggCursor.Close(ctx)
		var result struct {
			TotalAmount float64 `bson:"totalAmount"`
		}
		if aggCursor.Next(ctx) {
			_ = aggCursor.Decode(&result)
			totalSum = result.TotalAmount
		}
	}

	return bills, totalCount, totalSum, nil
}
