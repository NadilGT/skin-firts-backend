package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB_CreateStockMovement inserts a new immutable ledger entry.
// Call this after every quantity change: GRN, sale, transfer, reject, adjustment.
func DB_CreateStockMovement(m dto.StockMovementModel) error {
	if m.ID.IsZero() {
		m.ID = primitive.NewObjectID()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	_, err := dbConfigs.StockMovementCollection.InsertOne(context.Background(), m)
	return err
}

// DB_SearchStockMovements returns paginated movements filtered by query fields.
func DB_SearchStockMovements(query dto.SearchMovementQuery) ([]dto.StockMovementModel, int64, error) {
	ctx := context.Background()
	filter := bson.M{}

	if query.BatchId != "" {
		filter["batchId"] = query.BatchId
	}
	if query.MedicineId != "" {
		filter["medicineId"] = query.MedicineId
	}
	if query.BranchId != "" {
		filter["branchId"] = query.BranchId
	}
	if query.Type != "" {
		filter["type"] = query.Type
	}
	if query.ReferenceId != "" {
		filter["referenceId"] = query.ReferenceId
	}
	if query.ReferenceType != "" {
		filter["referenceType"] = query.ReferenceType
	}

	// Date range filter on createdAt
	if query.From != "" || query.To != "" {
		dateFilter := bson.M{}
		if query.From != "" {
			if t, err := time.Parse(time.RFC3339, query.From); err == nil {
				dateFilter["$gte"] = t
			}
		}
		if query.To != "" {
			if t, err := time.Parse(time.RFC3339, query.To); err == nil {
				dateFilter["$lte"] = t
			}
		}
		if len(dateFilter) > 0 {
			filter["createdAt"] = dateFilter
		}
	}

	total, err := dbConfigs.StockMovementCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 50
	}

	findOpts := options.Find().
		SetSkip(int64((query.Page-1)*query.Limit)).
		SetLimit(int64(query.Limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := dbConfigs.StockMovementCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var movements []dto.StockMovementModel
	if err = cursor.All(ctx, &movements); err != nil {
		return nil, 0, err
	}
	return movements, total, nil
}

// DB_GetMovementsByBatch returns all ledger entries for a given batchId, newest first.
func DB_GetMovementsByBatch(batchId string) ([]dto.StockMovementModel, error) {
	ctx := context.Background()
	filter := bson.M{"batchId": batchId}
	findOpts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := dbConfigs.StockMovementCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var movements []dto.StockMovementModel
	if err = cursor.All(ctx, &movements); err != nil {
		return nil, err
	}
	return movements, nil
}
