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

	// 1. Get Active Batches
	batchCursor, err := dbConfigs.MedicineBatchCollection.Find(ctx, bson.M{"status": "ACTIVE"})
	if err != nil {
		return nil, err
	}
	defer batchCursor.Close(ctx)

	var activeBatches []dto.MedicineBatch
	if err := batchCursor.All(ctx, &activeBatches); err != nil {
		return nil, err
	}

	batchMap := make(map[string]dto.MedicineBatch)
	var activeBatchIds []string
	for _, b := range activeBatches {
		batchMap[b.BatchId] = b
		activeBatchIds = append(activeBatchIds, b.BatchId)
	}

	if len(activeBatchIds) == 0 {
		return &dto.StockValuationResponse{BranchId: branchId, Items: []dto.StockValuationItem{}}, nil
	}

	// 2. Aggregate BranchStock for those active batches
	matchFilter := bson.M{
		"batchId":  bson.M{"$in": activeBatchIds},
		"quantity": bson.M{"$gt": 0},
	}
	if branchId != "" {
		matchFilter["branchId"] = branchId
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchFilter}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"medicineId": "$medicineId",
				"batchId":    "$batchId",
			},
			"totalQty": bson.M{"$sum": "$quantity"},
		}}},
	}

	stockCursor, err := dbConfigs.BranchStockCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer stockCursor.Close(ctx)

	type rawStockItem struct {
		ID struct {
			MedicineId string `bson:"medicineId"`
			BatchId    string `bson:"batchId"`
		} `bson:"_id"`
		TotalQty int `bson:"totalQty"`
	}

	var rawStocks []rawStockItem
	if err = stockCursor.All(ctx, &rawStocks); err != nil {
		return nil, err
	}

	// 3. Collect distinct Medicine IDs
	medIdMap := make(map[string]bool)
	for _, rs := range rawStocks {
		medIdMap[rs.ID.MedicineId] = true
	}

	var medicineIDs []string
	for id := range medIdMap {
		medicineIDs = append(medicineIDs, id)
	}

	nameMap, _ := DB_GetMedicineNamesByIDs(medicineIDs)

	// 4. Calculate totals per medicine
	type medTot struct {
		TotalQty       int
		TotalCostValue float64
		TotalSaleValue float64
		BatchCount     int
	}
	totals := make(map[string]*medTot)

	for _, rs := range rawStocks {
		medId := rs.ID.MedicineId
		if totals[medId] == nil {
			totals[medId] = &medTot{}
		}

		batch := batchMap[rs.ID.BatchId]
		costVal := float64(rs.TotalQty) * batch.BuyingPrice
		saleVal := float64(rs.TotalQty) * batch.SellingPrice

		totals[medId].TotalQty += rs.TotalQty
		totals[medId].TotalCostValue += costVal
		totals[medId].TotalSaleValue += saleVal
		totals[medId].BatchCount++
	}

	var items []dto.StockValuationItem
	var grandCost, grandSale float64

	for medId, tot := range totals {
		avgBuyPrice := 0.0
		if tot.TotalQty > 0 {
			avgBuyPrice = tot.TotalCostValue / float64(tot.TotalQty)
		}
		items = append(items, dto.StockValuationItem{
			MedicineID:     medId,
			MedicineName:   nameMap[medId],
			TotalQty:       tot.TotalQty,
			AvgBuyingPrice: avgBuyPrice,
			TotalCostValue: tot.TotalCostValue,
			TotalSaleValue: tot.TotalSaleValue,
		})
		grandCost += tot.TotalCostValue
		grandSale += tot.TotalSaleValue
	}

	// Sort items by name (simple bubble sort)
	for i := 0; i < len(items)-1; i++ {
		for j := 0; j < len(items)-i-1; j++ {
			if items[j].MedicineName > items[j+1].MedicineName {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
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

	// 1. Find batches that are expiring soon
	batchFilter := bson.M{
		"expiryDate": bson.M{
			"$lte": threshold,
			"$gte": time.Now(),
		},
	}
	batchCursor, err := dbConfigs.MedicineBatchCollection.Find(ctx, batchFilter)
	if err != nil {
		return nil, err
	}
	defer batchCursor.Close(ctx)

	var expiringBatches []dto.MedicineBatch
	if err = batchCursor.All(ctx, &expiringBatches); err != nil {
		return nil, err
	}

	if len(expiringBatches) == 0 {
		return []dto.ExpiryAlertItem{}, nil
	}

	batchMap := make(map[string]dto.MedicineBatch)
	var expiringBatchIds []string
	for _, b := range expiringBatches {
		batchMap[b.BatchId] = b
		expiringBatchIds = append(expiringBatchIds, b.BatchId)
	}

	// 2. Find branch stock for these batches
	stockFilter := bson.M{
		"batchId":  bson.M{"$in": expiringBatchIds},
		"quantity": bson.M{"$gt": 0},
	}
	if branchId != "" {
		stockFilter["branchId"] = branchId
	}

	stockCursor, err := dbConfigs.BranchStockCollection.Find(ctx, stockFilter)
	if err != nil {
		return nil, err
	}
	defer stockCursor.Close(ctx)

	var stocks []dto.BranchStock
	if err = stockCursor.All(ctx, &stocks); err != nil {
		return nil, err
	}

	if len(stocks) == 0 {
		return []dto.ExpiryAlertItem{}, nil
	}

	// 3. Gather medicine names
	medicineIDs := make([]string, 0)
	seen := map[string]bool{}
	for _, s := range stocks {
		if !seen[s.MedicineId] {
			medicineIDs = append(medicineIDs, s.MedicineId)
			seen[s.MedicineId] = true
		}
	}
	nameMap, _ := DB_GetMedicineNamesByIDs(medicineIDs)

	// 4. Construct response
	var alerts []dto.ExpiryAlertItem
	now := time.Now()
	for _, s := range stocks {
		batchInfo := batchMap[s.BatchId]
		daysLeft := int(batchInfo.ExpiryDate.Sub(now).Hours() / 24)
		alerts = append(alerts, dto.ExpiryAlertItem{
			BatchID:      batchInfo.BatchId,
			MedicineID:   s.MedicineId,
			MedicineName: nameMap[s.MedicineId],
			BatchNumber:  batchInfo.BatchNumber,
			ExpiryDate:   batchInfo.ExpiryDate,
			Quantity:     s.Quantity,
			DaysToExpiry: daysLeft,
			BranchId:     s.BranchId,
		})
	}

	// Sort by expiry date ascending (simple bubble sort)
	for i := 0; i < len(alerts)-1; i++ {
		for j := 0; j < len(alerts)-i-1; j++ {
			if alerts[j].ExpiryDate.After(alerts[j+1].ExpiryDate) {
				alerts[j], alerts[j+1] = alerts[j+1], alerts[j]
			}
		}
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
	if transfer.Status == "COMPLETED" {
		return fmt.Errorf("transfer is already COMPLETED and cannot be cancelled")
	}
	if transfer.Status == "CANCELLED" {
		return fmt.Errorf("transfer is already CANCELLED")
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

	// 1. Get Active Batches
	batchCursor, err := dbConfigs.MedicineBatchCollection.Find(ctx, bson.M{"status": "ACTIVE"})
	if err != nil {
		return nil, err
	}
	defer batchCursor.Close(ctx)

	var activeBatches []dto.MedicineBatch
	if err := batchCursor.All(ctx, &activeBatches); err != nil {
		return nil, err
	}

	var activeBatchIds []string
	for _, b := range activeBatches {
		activeBatchIds = append(activeBatchIds, b.BatchId)
	}

	if len(activeBatchIds) == 0 {
		return []dto.StockReportItem{}, nil
	}

	// 2. Aggregate BranchStock for those batches
	matchFilter := bson.M{
		"batchId": bson.M{"$in": activeBatchIds},
	}
	if branchId != "" {
		matchFilter["branchId"] = branchId
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchFilter}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":          "$medicineId",
			"totalQty":     bson.M{"$sum": "$quantity"},
			// totalBatches: number of unique batches with stock for this medicine
			"batchesSet":   bson.M{"$addToSet": "$batchId"},
		}}},
	}

	stockCursor, err := dbConfigs.BranchStockCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer stockCursor.Close(ctx)

	type rawItem struct {
		MedicineID string   `bson:"_id"`
		TotalQty   int      `bson:"totalQty"`
		BatchesSet []string `bson:"batchesSet"`
	}

	var rawItems []rawItem
	if err = stockCursor.All(ctx, &rawItems); err != nil {
		return nil, err
	}

	if len(rawItems) == 0 {
		return []dto.StockReportItem{}, nil
	}

	// 3. Collect distinct Medicine IDs
	var medicineIDs []string
	for _, r := range rawItems {
		medicineIDs = append(medicineIDs, r.MedicineID)
	}

	// Fetch full medicine details since we need Category, MinStockLevel, ReorderLevel
	medCursor, err := dbConfigs.MedicineCollection.Find(ctx, bson.M{"medicineid": bson.M{"$in": medicineIDs}})
	if err != nil {
		return nil, err
	}
	defer medCursor.Close(ctx)

	var medicines []dto.MedicineModel
	if err = medCursor.All(ctx, &medicines); err != nil {
		return nil, err
	}

	medMap := make(map[string]dto.MedicineModel)
	for _, m := range medicines {
		medMap[m.MedicineId] = m
	}

	// 4. Construct report
	var report []dto.StockReportItem
	for _, r := range rawItems {
		med := medMap[r.MedicineID]
		reorder := med.ReorderLevel
		if reorder == 0 {
			reorder = med.MinStockLevel
		}
		report = append(report, dto.StockReportItem{
			MedicineID:   r.MedicineID,
			MedicineName: med.Name,
			Category:     med.Category,
			TotalQty:     r.TotalQty,
			ReorderLevel: reorder,
			IsLowStock:   r.TotalQty <= reorder,
			TotalBatches: len(r.BatchesSet),
		})
	}

	// Sort by medicine name ascending
	for i := 0; i < len(report)-1; i++ {
		for j := 0; j < len(report)-i-1; j++ {
			if report[j].MedicineName > report[j+1].MedicineName {
				report[j], report[j+1] = report[j+1], report[j]
			}
		}
	}

	return report, nil
}
