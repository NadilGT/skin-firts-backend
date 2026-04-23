package dao

import (
	"context"
	"fmt"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


// ──────────────────────────────────────────────
//  Stock Valuation
// ──────────────────────────────────────────────

// DB_GetStockValuation returns a valuation of all active stock for a given branch.
// Pass branchId="" to get valuation across all branches.
func DB_GetStockValuation(branchId string) (*dto.StockValuationResponse, error) {
	ctx := context.Background()

	matchStage := bson.D{{Key: "$match", Value: bson.M{"status": "ACTIVE", "quantity": bson.M{"$gt": 0}}}}
	if branchId != "" {
		matchStage = bson.D{{Key: "$match", Value: bson.M{
			"status":   "ACTIVE",
			"quantity": bson.M{"$gt": 0},
			"branchId": branchId,
		}}}
	}

	pipeline := mongo.Pipeline{
		matchStage,
		bson.D{{Key: "$group", Value: bson.M{
			"_id":            "$medicineId",
			"totalQty":       bson.M{"$sum": "$quantity"},
			"avgBuyingPrice": bson.M{"$avg": "$buyingPrice"},
			"totalCostValue": bson.M{"$sum": bson.M{"$multiply": []interface{}{"$quantity", "$buyingPrice"}}},
			"totalSaleValue": bson.M{"$sum": bson.M{"$multiply": []interface{}{"$quantity", "$sellingPrice"}}},
		}}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "medicines",
			"localField":   "_id",
			"foreignField": "medicineid",
			"as":           "medicine",
		}}},
		bson.D{{Key: "$unwind", Value: bson.M{"path": "$medicine", "preserveNullAndEmptyArrays": true}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "medicine.name", Value: 1}}}},
	}

	cursor, err := dbConfigs.MedicineBatchCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type rawItem struct {
		MedicineID     string  `bson:"_id"`
		TotalQty       int     `bson:"totalQty"`
		AvgBuyingPrice float64 `bson:"avgBuyingPrice"`
		TotalCostValue float64 `bson:"totalCostValue"`
		TotalSaleValue float64 `bson:"totalSaleValue"`
		Medicine       struct {
			Name string `bson:"name"`
		} `bson:"medicine"`
	}

	var rawItems []rawItem
	if err = cursor.All(ctx, &rawItems); err != nil {
		return nil, err
	}

	var items []dto.StockValuationItem
	var grandCost, grandSale float64
	for _, r := range rawItems {
		items = append(items, dto.StockValuationItem{
			MedicineID:     r.MedicineID,
			MedicineName:   r.Medicine.Name,
			TotalQty:       r.TotalQty,
			AvgBuyingPrice: r.AvgBuyingPrice,
			TotalCostValue: r.TotalCostValue,
			TotalSaleValue: r.TotalSaleValue,
		})
		grandCost += r.TotalCostValue
		grandSale += r.TotalSaleValue
	}

	return &dto.StockValuationResponse{
		BranchId:       branchId,
		Items:          items,
		GrandCostValue: grandCost,
		GrandSaleValue: grandSale,
	}, nil
}

// ──────────────────────────────────────────────
//  Expiry Alerts
// ──────────────────────────────────────────────

// DB_GetExpiryAlerts returns batches expiring within 'days' days.
func DB_GetExpiryAlerts(branchId string, days int) ([]dto.ExpiryAlertItem, error) {
	ctx := context.Background()
	threshold := time.Now().AddDate(0, 0, days)

	pipeline := bson.A{
		// 1. Join branch_stock with medicine_batches
		bson.M{"$lookup": bson.M{
			"from":         "medicine_batches",
			"localField":   "batchId",
			"foreignField": "batchId",
			"as":           "batchInfo",
		}},
		bson.M{"$unwind": "$batchInfo"},
		// 2. Filter: quantity > 0, expiryDate <= threshold && >= now
		bson.M{"$match": bson.M{
			"quantity": bson.M{"$gt": 0},
			"batchInfo.expiryDate": bson.M{
				"$lte": threshold,
				"$gte": time.Now(),
			},
		}},
	}
	
	if branchId != "" {
		pipeline = append([]interface{}{
			bson.M{"$match": bson.M{"branchId": branchId}},
		}, pipeline...)
	}

	// 3. Sort by expiry date ascending
	pipeline = append(pipeline, bson.M{"$sort": bson.D{{Key: "batchInfo.expiryDate", Value: 1}}})

	cursor, err := dbConfigs.BranchStockCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// We decode into an anonymous struct temporarily since it's an aggregation result
	var results []struct {
		StockId    string           `bson:"stockId"`
		MedicineId string           `bson:"medicineId"`
		BranchId   string           `bson:"branchId"`
		Quantity   int              `bson:"quantity"`
		BatchInfo  dto.MedicineBatch `bson:"batchInfo"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Gather medicine names
	medicineIDs := make([]string, 0)
	seen := map[string]bool{}
	for _, r := range results {
		if !seen[r.MedicineId] {
			medicineIDs = append(medicineIDs, r.MedicineId)
			seen[r.MedicineId] = true
		}
	}
	nameMap, err := DB_GetMedicineNamesByIDs(medicineIDs)
	if err != nil {
		nameMap = map[string]string{}
	}

	var alerts []dto.ExpiryAlertItem
	now := time.Now()
	for _, r := range results {
		daysLeft := int(r.BatchInfo.ExpiryDate.Sub(now).Hours() / 24)
		alerts = append(alerts, dto.ExpiryAlertItem{
			BatchID:      r.BatchInfo.BatchId,
			MedicineID:   r.MedicineId,
			MedicineName: nameMap[r.MedicineId],
			BatchNumber:  r.BatchInfo.BatchNumber,
			ExpiryDate:   r.BatchInfo.ExpiryDate,
			Quantity:     r.Quantity,
			DaysToExpiry: daysLeft,
			BranchId:     r.BranchId,
		})
	}
	return alerts, nil
}

// ──────────────────────────────────────────────
//  Stock Transfer
// ──────────────────────────────────────────────

// DB_ReserveSpecificStock reserves a specific quantity of stock from a given BranchStock record.
func DB_ReserveSpecificStock(stockId string, branchId string, requiredQty int) error {
	filter := bson.M{
		"stockId": stockId,
		"branchId": branchId,
		"$expr": bson.M{
			"$gte": bson.A{
				bson.M{"$subtract": bson.A{"$quantity", "$reservedQuantity"}},
				requiredQty,
			},
		},
	}
	update := bson.M{
		"$inc": bson.M{"reservedQuantity": requiredQty},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	res, err := dbConfigs.BranchStockCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return fmt.Errorf("insufficient available stock or stock record not found for stockId: %s", stockId)
	}
	return nil
}

// DB_RevertTransferStockReservation removes the reservation for a list of transfer items.
func DB_RevertTransferStockReservation(items []dto.TransferItem, branchId string) error {
	for _, item := range items {
		if item.StockId != "" {
			filter := bson.M{"stockId": item.StockId, "branchId": branchId, "reservedQuantity": bson.M{"$gte": item.Quantity}}
			update := bson.M{
				"$inc": bson.M{"reservedQuantity": -item.Quantity},
				"$set": bson.M{"updatedAt": time.Now()},
			}
			_ = dbConfigs.BranchStockCollection.FindOneAndUpdate(context.Background(), filter, update)
		}
	}
	return nil
}

func DB_CreateStockTransfer(transfer dto.StockTransferModel) error {
	_, err := dbConfigs.StockTransferCollection.InsertOne(context.Background(), transfer)
	return err
}

func DB_GetStockTransferByTransferID(transferId string) (*dto.StockTransferModel, error) {
	var t dto.StockTransferModel
	err := dbConfigs.StockTransferCollection.FindOne(context.Background(), bson.M{"transferId": transferId}).Decode(&t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func DB_SearchStockTransfers(query dto.SearchTransferQuery) ([]dto.StockTransferModel, int64, error) {
	ctx := context.Background()
	filter := bson.M{}
	if query.FromBranchId != "" {
		filter["fromBranchId"] = query.FromBranchId
	}
	if query.ToBranchId != "" {
		filter["toBranchId"] = query.ToBranchId
	}
	if query.Status != "" {
		filter["status"] = query.Status
	}
	total, err := dbConfigs.StockTransferCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	findOpts := options.Find().
		SetSkip(int64((query.Page - 1) * query.Limit)).
		SetLimit(int64(query.Limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := dbConfigs.StockTransferCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var transfers []dto.StockTransferModel
	if err = cursor.All(ctx, &transfers); err != nil {
		return nil, 0, err
	}
	return transfers, total, nil
}

// DB_ApproveStockTransfer transitions a transfer from PENDING → APPROVED
// and creates an Approval record so execution can gate on it.
func DB_ApproveStockTransfer(transferId string, approvedBy string) error {
	ctx := context.Background()

	transfer, err := DB_GetStockTransferByTransferID(transferId)
	if err != nil {
		return err
	}
	if transfer.Status != "PENDING" {
		return fmt.Errorf("transfer is already %s", transfer.Status)
	}

	// Update status to APPROVED
	_, err = dbConfigs.StockTransferCollection.UpdateOne(ctx,
		bson.M{"transferId": transferId},
		bson.M{"$set": bson.M{"status": "APPROVED", "updatedAt": time.Now()}})
	if err != nil {
		return err
	}

	// Create approval record
	approvalId, err := GenerateId(ctx, "approvals", "APR")
	if err != nil {
		return fmt.Errorf("transfer approved but failed to generate approval id: %v", err)
	}
	return DB_CreateApproval(dto.ApprovalModel{
		ID:            primitive.NewObjectID(),
		ApprovalId:    approvalId,
		ReferenceType: dto.ApprovalRefTransfer,
		ReferenceId:   transfer.TransferId,
		Status:        dto.ApprovalApproved,
		ApprovedBy:    approvedBy,
		ApprovedAt:    time.Now(),
		CreatedAt:     time.Now(),
	})
}

// DB_CompleteStockTransfer executes an APPROVED transfer:
//  1. Deducts qty from each source batch (TRANSFER_OUT movement)
//  2. Creates a new batch in the target branch (TRANSFER_IN movement)
//  3. Marks the transfer as COMPLETED
func DB_CompleteStockTransfer(transferId string) error {
	ctx := context.Background()

	transfer, err := DB_GetStockTransferByTransferID(transferId)
	if err != nil {
		return err
	}

	// ── Approval gate: transfer must be APPROVED before execution ──
	if transfer.Status != "APPROVED" {
		if transfer.Status == "PENDING" {
			return fmt.Errorf("transfer must be APPROVED before completion; call /stock-transfers/%s/approve first", transferId)
		}
		return fmt.Errorf("transfer is already %s", transfer.Status)
	}

	for _, item := range transfer.Items {
		// 1. Atomically deduct from source BranchStock
		// Note: We decrement BOTH quantity and reservedQuantity because it was reserved on creation
		var srcStock dto.BranchStock
		err := dbConfigs.BranchStockCollection.FindOne(ctx, bson.M{"stockId": item.StockId, "branchId": transfer.FromBranchId}).Decode(&srcStock)
		if err != nil {
			return fmt.Errorf("invalid stockId %s in source branch: %v", item.StockId, err)
		}

		filter := bson.M{"_id": srcStock.ID, "quantity": bson.M{"$gte": item.Quantity}}
		update := bson.M{
			"$inc": bson.M{"quantity": -item.Quantity, "reservedQuantity": -item.Quantity},
			"$set": bson.M{"updatedAt": time.Now()},
		}
		opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
		if err := dbConfigs.BranchStockCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&srcStock); err != nil {
			return fmt.Errorf("insufficient stock in source branch for stockId %s", item.StockId)
		}

		// ── Write TRANSFER_OUT movement on source branch ──
		outMovId, _ := GenerateId(ctx, "stock_movements", "MOV")
		_ = DB_CreateStockMovement(dto.StockMovementModel{
			ID:            primitive.NewObjectID(),
			MovementId:    outMovId,
			BatchId:       item.BatchId,
			MedicineId:    item.MedicineID,
			BranchId:      transfer.FromBranchId,
			Type:          dto.MovementTransferOut,
			Quantity:      item.Quantity,
			ReferenceId:   transfer.TransferId,
			ReferenceType: "TRANSFER",
			Notes:         fmt.Sprintf("Transfer out to branch %s", transfer.ToBranchId),
			CreatedAt:     time.Now(),
		})

		// 2. Add stock to target BranchStock (Upsert)
		targetStockId, _ := GenerateId(ctx, "branch_stock", "STK")
		targetFilter := bson.M{"batchId": item.BatchId, "branchId": transfer.ToBranchId}
		targetUpdate := bson.M{
			"$inc": bson.M{"quantity": item.Quantity},
			"$setOnInsert": bson.M{
				"_id":              primitive.NewObjectID(),
				"stockId":          targetStockId,
				"medicineId":       item.MedicineID,
				"reservedQuantity": 0,
			},
			"$set": bson.M{"updatedAt": time.Now()},
		}
		upsertOpt := options.Update().SetUpsert(true)
		if _, err := dbConfigs.BranchStockCollection.UpdateOne(ctx, targetFilter, targetUpdate, upsertOpt); err != nil {
			return fmt.Errorf("failed to add stock to target branch: %v", err)
		}

		// ── Write TRANSFER_IN movement on destination branch ──
		inMovId, _ := GenerateId(ctx, "stock_movements", "MOV")
		_ = DB_CreateStockMovement(dto.StockMovementModel{
			ID:            primitive.NewObjectID(),
			MovementId:    inMovId,
			BatchId:       item.BatchId,
			MedicineId:    item.MedicineID,
			BranchId:      transfer.ToBranchId,
			Type:          dto.MovementTransferIn,
			Quantity:      item.Quantity,
			ReferenceId:   transfer.TransferId,
			ReferenceType: "TRANSFER",
			Notes:         fmt.Sprintf("Transfer in from branch %s", transfer.FromBranchId),
			CreatedAt:     time.Now(),
		})
	}


	// Mark transfer as COMPLETED
	_, err = dbConfigs.StockTransferCollection.UpdateOne(ctx,
		bson.M{"transferId": transferId},
		bson.M{"$set": bson.M{"status": "COMPLETED", "updatedAt": time.Now()}})
	return err
}

func DB_CancelStockTransfer(transferId string) error {
	ctx := context.Background()

	transfer, err := DB_GetStockTransferByTransferID(transferId)
	if err != nil {
		return err
	}

	// Revert stock reservation if the transfer was PENDING or APPROVED
	if transfer.Status == "PENDING" || transfer.Status == "APPROVED" {
		_ = DB_RevertTransferStockReservation(transfer.Items, transfer.FromBranchId)
	}

	_, err = dbConfigs.StockTransferCollection.UpdateOne(
		ctx,
		bson.M{"transferId": transferId},
		bson.M{"$set": bson.M{"status": "CANCELLED", "updatedAt": time.Now()}},
	)
	return err
}

// ──────────────────────────────────────────────
//  Stock Report
// ──────────────────────────────────────────────

func DB_GetStockReport(branchId string) ([]dto.StockReportItem, error) {
	ctx := context.Background()

	matchStage := bson.D{{Key: "$match", Value: bson.M{"status": "ACTIVE"}}}
	if branchId != "" {
		matchStage = bson.D{{Key: "$match", Value: bson.M{"status": "ACTIVE", "branchId": branchId}}}
	}

	pipeline := mongo.Pipeline{
		matchStage,
		bson.D{{Key: "$group", Value: bson.M{
			"_id":          "$medicineId",
			"totalQty":     bson.M{"$sum": "$quantity"},
			"totalBatches": bson.M{"$sum": 1},
		}}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "medicines",
			"localField":   "_id",
			"foreignField": "medicineid",
			"as":           "medicine",
		}}},
		bson.D{{Key: "$unwind", Value: bson.M{"path": "$medicine", "preserveNullAndEmptyArrays": true}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "medicine.name", Value: 1}}}},
	}

	cursor, err := dbConfigs.MedicineBatchCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type rawItem struct {
		MedicineID   string `bson:"_id"`
		TotalQty     int    `bson:"totalQty"`
		TotalBatches int    `bson:"totalBatches"`
		Medicine     struct {
			Name         string `bson:"name"`
			Category     string `bson:"category"`
			MinStockLevel int   `bson:"minStockLevel"`
			ReorderLevel int    `bson:"reorderLevel"`
		} `bson:"medicine"`
	}

	var rawItems []rawItem
	if err = cursor.All(ctx, &rawItems); err != nil {
		return nil, err
	}

	var report []dto.StockReportItem
	for _, r := range rawItems {
		reorder := r.Medicine.ReorderLevel
		if reorder == 0 {
			reorder = r.Medicine.MinStockLevel
		}
		report = append(report, dto.StockReportItem{
			MedicineID:   r.MedicineID,
			MedicineName: r.Medicine.Name,
			Category:     r.Medicine.Category,
			TotalQty:     r.TotalQty,
			ReorderLevel: reorder,
			IsLowStock:   r.TotalQty <= reorder,
			TotalBatches: r.TotalBatches,
		})
	}
	return report, nil
}
