package dao

import (
	"context"
	"fmt"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DB_CreateStockAdjustment(adj dto.StockAdjustmentModel) error {
	if adj.ID.IsZero() {
		adj.ID = primitive.NewObjectID()
	}
	_, err := dbConfigs.StockAdjustmentCollection.InsertOne(context.Background(), adj)
	return err
}

func DB_GetStockAdjustmentByID(id primitive.ObjectID) (*dto.StockAdjustmentModel, error) {
	var a dto.StockAdjustmentModel
	err := dbConfigs.StockAdjustmentCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func DB_ApproveStockAdjustment(id primitive.ObjectID, approvedBy string, notes string) error {
	ctx := context.Background()
	adj, err := DB_GetStockAdjustmentByID(id)
	if err != nil {
		return err
	}
	if adj.Status != "PENDING" {
		return fmt.Errorf("adjustment is already %s", adj.Status)
	}

	_, err = dbConfigs.StockAdjustmentCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"status":     "APPROVED",
			"approvedBy": approvedBy,
			"approvedAt": time.Now(),
			"notes":      notes,
			"updatedAt":  time.Now(),
		}},
	)
	if err != nil {
		return err
	}

	// Create Approval record
	approvalId, _ := GenerateId(ctx, "approvals", "APR")
	_ = DB_CreateApproval(dto.ApprovalModel{
		ID:            primitive.NewObjectID(),
		ApprovalId:    approvalId,
		ReferenceType: dto.ApprovalRefTransfer, // Reusing Transfer enum logically or add new. Let's just use string manually here.
		ReferenceId:   adj.AdjustmentId,
		Status:        dto.ApprovalApproved,
		ApprovedBy:    approvedBy,
		ApprovedAt:    time.Now(),
		Notes:         notes,
		CreatedAt:     time.Now(),
	})
	return nil
}

func DB_ExecuteStockAdjustment(id primitive.ObjectID, executedBy string) error {
	ctx := context.Background()

	adj, err := DB_GetStockAdjustmentByID(id)
	if err != nil {
		return err
	}
	if adj.Status != "APPROVED" {
		return fmt.Errorf("adjustment must be APPROVED before execution")
	}

	// 1. Check expiration from global MedicineBatch
	var batch dto.MedicineBatch
	err = dbConfigs.MedicineBatchCollection.FindOne(ctx, bson.M{"batchId": adj.BatchId}).Decode(&batch)
	if err != nil {
		return fmt.Errorf("batch not found")
	}
	if time.Now().After(batch.ExpiryDate) {
		return fmt.Errorf("cannot adjust an expired batch")
	}

	// 2. Load BranchStock
	var stock dto.BranchStock
	err = dbConfigs.BranchStockCollection.FindOne(ctx, bson.M{"stockId": adj.StockId, "branchId": adj.BranchId}).Decode(&stock)
	if err != nil {
		return fmt.Errorf("stock record not found for branch")
	}

	updateMod := 0
	if adj.Type == "ADJUSTMENT_IN" {
		updateMod = adj.Quantity
	} else if adj.Type == "ADJUSTMENT_OUT" {
		// Calculate available mathematically without blocking reservations
		available := stock.Quantity - stock.ReservedQuantity
		if available < adj.Quantity {
			return fmt.Errorf("insufficient stock for OUT adjustment")
		}
		updateMod = -adj.Quantity
	} else {
		return fmt.Errorf("invalid adjustment type")
	}

	update := bson.M{
		"$inc": bson.M{"quantity": updateMod},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated dto.BranchStock
	err = dbConfigs.BranchStockCollection.FindOneAndUpdate(ctx, bson.M{"_id": stock.ID}, update, opts).Decode(&updated)
	if err != nil {
		return err
	}

	// 3. Write Movement
	movementId, _ := GenerateId(ctx, "stock_movements", "MOV")
	_ = DB_CreateStockMovement(dto.StockMovementModel{
		ID:            primitive.NewObjectID(),
		MovementId:    movementId,
		BatchId:       adj.BatchId,
		MedicineId:    adj.MedicineId,
		BranchId:      adj.BranchId,
		Type:          dto.MovementAdjustment,
		Quantity:      adj.Quantity,
		ReferenceId:   adj.AdjustmentId,
		ReferenceType: "ADJUSTMENT",
		Notes:         fmt.Sprintf("Type: %s, Reason: %s", adj.Type, adj.Reason),
		CreatedBy:     executedBy,
		CreatedAt:     time.Now(),
	})

	_, err = dbConfigs.StockAdjustmentCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": "COMPLETED", "updatedAt": time.Now()}},
	)
	return err
}

func DB_SearchStockAdjustments(query dto.SearchAdjustmentQuery) ([]dto.StockAdjustmentModel, int64, error) {
	ctx := context.Background()
	filter := bson.M{}

	if query.BatchId != "" {
		filter["batchId"] = query.BatchId
	}
	if query.BranchId != "" {
		filter["branchId"] = query.BranchId
	}
	if query.Status != "" {
		filter["status"] = query.Status
	}
	total, err := dbConfigs.StockAdjustmentCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
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

	cursor, err := dbConfigs.StockAdjustmentCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var records []dto.StockAdjustmentModel
	if err = cursor.All(ctx, &records); err != nil {
		return nil, 0, err
	}
	return records, total, nil
}
