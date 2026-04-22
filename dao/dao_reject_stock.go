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

// DB_CreateRejectStock creates a new reject stock request in PENDING status.
func DB_CreateRejectStock(r dto.RejectStockModel) error {
	if r.ID.IsZero() {
		r.ID = primitive.NewObjectID()
	}
	_, err := dbConfigs.RejectStockCollection.InsertOne(context.Background(), r)
	return err
}

// DB_GetRejectStockByID fetches a reject stock record by its MongoDB ObjectID.
func DB_GetRejectStockByID(id primitive.ObjectID) (*dto.RejectStockModel, error) {
	var r dto.RejectStockModel
	err := dbConfigs.RejectStockCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// DB_ApproveRejectStock transitions a reject record from PENDING → APPROVED.
// It also creates an Approval record for the audit trail.
func DB_ApproveRejectStock(id primitive.ObjectID, approvedBy string, notes string) error {
	r, err := DB_GetRejectStockByID(id)
	if err != nil {
		return err
	}
	if r.Status != "PENDING" {
		return fmt.Errorf("reject stock is already %s", r.Status)
	}

	now := time.Now()
	_, err = dbConfigs.RejectStockCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"status":     "APPROVED",
			"approvedBy": approvedBy,
			"approvedAt": now,
			"notes":      notes,
			"updatedAt":  now,
		}},
	)
	return err
}

// DB_ExecuteRejectStock transitions APPROVED → COMPLETED, deducts batch quantity,
// and records a REJECT StockMovement in the ledger.
func DB_ExecuteRejectStock(id primitive.ObjectID, executedBy string) error {
	ctx := context.Background()

	r, err := DB_GetRejectStockByID(id)
	if err != nil {
		return err
	}
	if r.Status != "APPROVED" {
		return fmt.Errorf("reject stock must be APPROVED before execution (current: %s)", r.Status)
	}

	// Parse batchId as ObjectID
	batchObjID, err := primitive.ObjectIDFromHex(r.BatchId)
	if err != nil {
		return fmt.Errorf("invalid batchId %s: %v", r.BatchId, err)
	}

	// Atomically deduct quantity from batch — fail if not enough stock
	filter := bson.M{"_id": batchObjID, "quantity": bson.M{"$gte": r.Quantity}}
	update := bson.M{"$inc": bson.M{"quantity": -r.Quantity}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedBatch dto.MedicineBatchModel
	if err := dbConfigs.MedicineBatchCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedBatch); err != nil {
		return fmt.Errorf("insufficient stock in batch %s or batch not found: %v", r.BatchId, err)
	}

	// Write REJECT StockMovement to the ledger
	movementId, err := GenerateId(ctx, "stock_movements", "MOV")
	if err != nil {
		return fmt.Errorf("failed to generate movement id: %v", err)
	}
	movement := dto.StockMovementModel{
		ID:            primitive.NewObjectID(),
		MovementId:    movementId,
		BatchId:       r.BatchId,
		MedicineId:    r.MedicineId,
		BranchId:      r.BranchId,
		Type:          dto.MovementReject,
		Quantity:      r.Quantity,
		ReferenceId:   r.RejectId,
		ReferenceType: "REJECT",
		Notes:         fmt.Sprintf("Reject type: %s — %s", r.Type, r.Reason),
		CreatedBy:     executedBy,
		CreatedAt:     time.Now(),
	}
	if err := DB_CreateStockMovement(movement); err != nil {
		return fmt.Errorf("stock deducted but failed to write movement: %v", err)
	}

	// Mark reject as COMPLETED
	_, err = dbConfigs.RejectStockCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": "COMPLETED", "updatedAt": time.Now()}},
	)
	return err
}

// DB_SearchRejectStock returns paginated reject stock records.
func DB_SearchRejectStock(query dto.SearchRejectQuery) ([]dto.RejectStockModel, int64, error) {
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
	if query.Status != "" {
		filter["status"] = query.Status
	}
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

	total, err := dbConfigs.RejectStockCollection.CountDocuments(ctx, filter)
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
		SetSkip(int64((query.Page-1)*query.Limit)).
		SetLimit(int64(query.Limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := dbConfigs.RejectStockCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var records []dto.RejectStockModel
	if err = cursor.All(ctx, &records); err != nil {
		return nil, 0, err
	}
	return records, total, nil
}
