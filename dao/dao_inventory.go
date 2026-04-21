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

// DB_GetExpiryAlerts returns batches expiring within `days` days.
func DB_GetExpiryAlerts(branchId string, days int) ([]dto.ExpiryAlertItem, error) {
	ctx := context.Background()

	if days <= 0 {
		days = 90
	}
	threshold := time.Now().AddDate(0, 0, days)

	filter := bson.M{
		"status":   "ACTIVE",
		"quantity": bson.M{"$gt": 0},
		"expiryDate": bson.M{
			"$lte": threshold,
			"$gte": time.Now(),
		},
	}
	if branchId != "" {
		filter["branchId"] = branchId
	}

	findOpts := options.Find().SetSort(bson.D{{Key: "expiryDate", Value: 1}})
	cursor, err := dbConfigs.MedicineBatchCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var batches []dto.MedicineBatchModel
	if err = cursor.All(ctx, &batches); err != nil {
		return nil, err
	}

	// Gather medicine names
	medicineIDs := make([]string, 0)
	seen := map[string]bool{}
	for _, b := range batches {
		if !seen[b.MedicineID] {
			medicineIDs = append(medicineIDs, b.MedicineID)
			seen[b.MedicineID] = true
		}
	}
	nameMap, err := DB_GetMedicineNamesByIDs(medicineIDs)
	if err != nil {
		nameMap = map[string]string{}
	}

	var alerts []dto.ExpiryAlertItem
	now := time.Now()
	for _, b := range batches {
		daysLeft := int(b.ExpiryDate.Sub(now).Hours() / 24)
		alerts = append(alerts, dto.ExpiryAlertItem{
			BatchID:      b.ID.Hex(),
			MedicineID:   b.MedicineID,
			MedicineName: nameMap[b.MedicineID],
			BatchNumber:  b.BatchNumber,
			ExpiryDate:   b.ExpiryDate,
			Quantity:     b.Quantity,
			DaysToExpiry: daysLeft,
			BranchId:     b.BranchId,
		})
	}
	return alerts, nil
}

// ──────────────────────────────────────────────
//  Stock Transfer
// ──────────────────────────────────────────────

func DB_CreateStockTransfer(transfer dto.StockTransferModel) error {
	_, err := dbConfigs.StockTransferCollection.InsertOne(context.Background(), transfer)
	return err
}

func DB_GetStockTransferByID(id primitive.ObjectID) (*dto.StockTransferModel, error) {
	var t dto.StockTransferModel
	err := dbConfigs.StockTransferCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&t)
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

// DB_CompleteStockTransfer deducts stock from source batches and creates new batches in the target branch.
func DB_CompleteStockTransfer(transferID primitive.ObjectID) error {
	ctx := context.Background()

	transfer, err := DB_GetStockTransferByID(transferID)
	if err != nil {
		return err
	}
	if transfer.Status != "PENDING" {
		return fmt.Errorf("transfer is already %s", transfer.Status)
	}

	for _, item := range transfer.Items {
		batchObjID, err := primitive.ObjectIDFromHex(item.BatchId)
		if err != nil {
			return fmt.Errorf("invalid batchId %s: %v", item.BatchId, err)
		}

		// Deduct from source batch atomically
		filter := bson.M{"_id": batchObjID, "quantity": bson.M{"$gte": item.Quantity}}
		update := bson.M{"$inc": bson.M{"quantity": -item.Quantity}}
		opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
		var srcBatch dto.MedicineBatchModel
		if err := dbConfigs.MedicineBatchCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&srcBatch); err != nil {
			return fmt.Errorf("insufficient stock in batch %s: %v", item.BatchId, err)
		}
		// Mark out-of-stock if depleted
		if srcBatch.Quantity == 0 {
			_, _ = dbConfigs.MedicineBatchCollection.UpdateOne(ctx,
				bson.M{"_id": batchObjID},
				bson.M{"$set": bson.M{"status": "OUT_OF_STOCK"}})
		}

		// Create a new batch in the target branch
		newBatchId, err := GenerateId(ctx, "medicine_batches", "BAT")
		if err != nil {
			return err
		}
		newBatch := dto.MedicineBatchModel{
			ID:              primitive.NewObjectID(),
			MedicineBatchId: newBatchId,
			MedicineID:      item.MedicineID,
			Quantity:        item.Quantity,
			ExpiryDate:      srcBatch.ExpiryDate,
			BuyingPrice:     srcBatch.BuyingPrice,
			SellingPrice:    srcBatch.SellingPrice,
			Status:          "ACTIVE",
			BatchNumber:     item.BatchNumber,
			BranchId:        transfer.ToBranchId,
			Notes:           fmt.Sprintf("Transferred from branch %s via transfer %s", transfer.FromBranchId, transfer.TransferId),
			CreatedAt:       time.Now(),
		}
		if _, err := dbConfigs.MedicineBatchCollection.InsertOne(ctx, newBatch); err != nil {
			return err
		}
	}

	// Mark transfer as completed
	_, err = dbConfigs.StockTransferCollection.UpdateOne(ctx,
		bson.M{"_id": transferID},
		bson.M{"$set": bson.M{"status": "COMPLETED", "updatedAt": time.Now()}})
	return err
}

func DB_CancelStockTransfer(id primitive.ObjectID) error {
	_, err := dbConfigs.StockTransferCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": id},
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
