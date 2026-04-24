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


func DB_CreateMedicine(medicine dto.MedicineModel) error {
	_, err := dbConfigs.MedicineCollection.InsertOne(context.Background(), medicine)
	if err != nil {
		return err
	}
	return nil
}

// DB_ExpireOldBills finds and cancels bills that have been PENDING for > 30 mins
func DB_ExpireOldBills() {
	ctx := context.Background()
	timeoutThreshold := time.Now().Add(-30 * time.Minute)

	filter := bson.M{
		"status":    "PENDING",
		"createdAt": bson.M{"$lt": timeoutThreshold},
	}

	cursor, err := dbConfigs.BillCollection.Find(ctx, filter)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	var expiredBills []dto.BillModel
	if err = cursor.All(ctx, &expiredBills); err != nil {
		return
	}

	for _, bill := range expiredBills {
		// Logically this is exactly what CancelBill does
		DB_RevertStockReservation(bill.Items)
		_ = DB_UpdateBillStatus(bill.BillId, "CANCELLED", bill.BranchId)
		fmt.Printf("[WMS-CRON] Auto-expired Bill %s\n", bill.BillId)
	}
}

// StartBillExpiryCron initializes a background ticker to clear stale reservations
func StartBillExpiryCron() {
	ticker := time.NewTicker(15 * time.Minute)
	go func() {
		for range ticker.C {
			DB_ExpireOldBills()
		}
	}()
}

func DB_SearchMedicines(query dto.SearchMedicineQuery) ([]dto.MedicineModel, int64, error) {
	ctx := context.Background()
	
	// Build filter
	filter := bson.M{}
	
	if query.SearchTerm != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": query.SearchTerm, "$options": "i"}},
			{"genericName": bson.M{"$regex": query.SearchTerm, "$options": "i"}},
			{"manufacturer": bson.M{"$regex": query.SearchTerm, "$options": "i"}},
		}
	}
	
	if query.Category != "" {
		filter["category"] = query.Category
	}
	
	if query.Form != "" {
		filter["form"] = query.Form
	}
	
	if query.Status != "" {
		filter["status"] = query.Status
	}
	
	// Get total count
	total, err := dbConfigs.MedicineCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	
	// Set options for pagination
	findOptions := options.Find()
	findOptions.SetSkip(int64((query.Page - 1) * query.Limit))
	findOptions.SetLimit(int64(query.Limit))
	findOptions.SetSort(bson.D{{Key: "name", Value: 1}})
	
	cursor, err := dbConfigs.MedicineCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	
	var medicines []dto.MedicineModel
	if err = cursor.All(ctx, &medicines); err != nil {
		return nil, 0, err
	}
	
	return medicines, total, nil
}

func DB_GetMedicineByID(id primitive.ObjectID) (*dto.MedicineModel, error) {
	var medicine dto.MedicineModel
	err := dbConfigs.MedicineCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&medicine)
	if err != nil {
		return nil, err
	}
	return &medicine, nil
}

func DB_UpdateMedicine(medicine dto.MedicineModel) error {
	filter := bson.M{"_id": medicine.ID}
	update := bson.M{
		"$set": bson.M{
			"name":                 medicine.Name,
			"genericName":          medicine.GenericName,
			"manufacturer":         medicine.Manufacturer,
			"category":             medicine.Category,
			"dosage":               medicine.Dosage,
			"form":                 medicine.Form,
			"strength":             medicine.Strength,
			"minStockLevel":        medicine.MinStockLevel,
			"barcode":              medicine.Barcode,
			"supplierId":           medicine.SupplierId,
			"reorderLevel":         medicine.ReorderLevel,
			"description":          medicine.Description,
			"sideEffects":          medicine.SideEffects,
			"contraindications":    medicine.Contraindications,
			"prescriptionRequired": medicine.PrescriptionRequired,
			"status":               medicine.Status,
			"updatedAt":            time.Now(),
		},
	}
	
	_, err := dbConfigs.MedicineCollection.UpdateOne(context.Background(), filter, update)
	return err
}

func DB_DeleteMedicine(id primitive.ObjectID) error {
	_, err := dbConfigs.MedicineCollection.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}

func DB_GetMedicineByBarcode(barcode string) (*dto.MedicineModel, error) {
	var medicine dto.MedicineModel
	err := dbConfigs.MedicineCollection.FindOne(context.Background(), bson.M{"barcode": barcode}).Decode(&medicine)
	if err != nil {
		return nil, err
	}
	return &medicine, nil
}

func DB_GetLowStockMedicines() ([]dto.MedicineModel, error) {
	ctx := context.Background()
	
	// 1. Get all active batches
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

	// 2. Aggregate BranchStock for these active batches to get total active stock per medicine
	stockMap := make(map[string]int)
	if len(activeBatchIds) > 0 {
		pipeline := mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{
				"batchId": bson.M{"$in": activeBatchIds},
			}}},
			bson.D{{Key: "$group", Value: bson.M{
				"_id": "$medicineId",
				"totalStock": bson.M{"$sum": "$quantity"},
			}}},
		}
		
		stockCursor, err := dbConfigs.BranchStockCollection.Aggregate(ctx, pipeline)
		if err == nil {
			defer stockCursor.Close(ctx)
			type rawStock struct {
				MedicineId string `bson:"_id"`
				TotalStock int    `bson:"totalStock"`
			}
			var rawStocks []rawStock
			if stockCursor.All(ctx, &rawStocks) == nil {
				for _, rs := range rawStocks {
					stockMap[rs.MedicineId] = rs.TotalStock
				}
			}
		}
	}

	// 3. Fetch all medicines and filter those with stock <= minStockLevel
	medCursor, err := dbConfigs.MedicineCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer medCursor.Close(ctx)
	
	var allMedicines []dto.MedicineModel
	if err = medCursor.All(ctx, &allMedicines); err != nil {
		return nil, err
	}
	
	var lowStockMedicines []dto.MedicineModel
	for _, m := range allMedicines {
		totalStock := stockMap[m.MedicineId]
		if totalStock <= m.MinStockLevel {
			lowStockMedicines = append(lowStockMedicines, m)
		}
	}
	
	return lowStockMedicines, nil
}

// ─────────────────────────────────────────────────────
//  MedicineBatch (global) operations
// ─────────────────────────────────────────────────────

// DB_CreateMedicineBatch inserts a global batch record (no qty, no branch).
func DB_CreateMedicineBatch(batch dto.MedicineBatch) error {
	_, err := dbConfigs.MedicineBatchCollection.InsertOne(context.Background(), batch)
	return err
}

// DB_CreateBranchStock inserts a branch-specific stock record.
func DB_CreateBranchStock(stock dto.BranchStock) error {
	_, err := dbConfigs.BranchStockCollection.InsertOne(context.Background(), stock)
	return err
}

// DB_GetBatchesByMedicineID returns all global batches for a medicine.
func DB_GetBatchesByMedicineID(medicineID string) ([]dto.MedicineBatch, error) {
	ctx := context.Background()
	cursor, err := dbConfigs.MedicineBatchCollection.Find(ctx, bson.M{"medicineId": medicineID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var batches []dto.MedicineBatch
	if err = cursor.All(ctx, &batches); err != nil {
		return nil, err
	}
	return batches, nil
}

// DB_GetBranchStockByBatch returns a branch's stock record for a specific batch.
func DB_GetBranchStockByBatch(batchId, branchId string) (*dto.BranchStock, error) {
	var s dto.BranchStock
	err := dbConfigs.BranchStockCollection.FindOne(context.Background(),
		bson.M{"batchId": batchId, "branchId": branchId}).Decode(&s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// DB_GetAvailableBatchesFEFO performs an aggregation JOIN in Go layer:
//   medicine_batches (for expiryDate + prices) + branch_stock (for branchId + available qty)
// Returns BranchStockView sorted by expiryDate ASC (First Expiring First Out).
func DB_GetAvailableBatchesFEFO(medicineID, branchId string) ([]dto.BranchStockView, error) {
	ctx := context.Background()

	// 1. Get non-expired, non-blocked batches for this medicine
	batchFilter := bson.M{
		"medicineId": medicineID,
		"expiryDate": bson.M{"$gt": time.Now()},
		"status":     bson.M{"$ne": "BLOCKED"},
	}
	batchCursor, err := dbConfigs.MedicineBatchCollection.Find(ctx, batchFilter)
	if err != nil {
		return nil, err
	}
	defer batchCursor.Close(ctx)

	var validBatches []dto.MedicineBatch
	if err = batchCursor.All(ctx, &validBatches); err != nil {
		return nil, err
	}

	if len(validBatches) == 0 {
		return []dto.BranchStockView{}, nil
	}

	batchMap := make(map[string]dto.MedicineBatch)
	var validBatchIds []string
	for _, b := range validBatches {
		batchMap[b.BatchId] = b
		validBatchIds = append(validBatchIds, b.BatchId)
	}

	// 2. Find branch stock for these valid batches where available > 0
	stockFilter := bson.M{
		"medicineId": medicineID,
		"branchId":   branchId,
		"batchId":    bson.M{"$in": validBatchIds},
		"$expr": bson.M{
			"$gt": bson.A{
				bson.M{"$subtract": bson.A{"$quantity", "$reservedQuantity"}},
				0,
			},
		},
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

	// 3. Construct BranchStockView
	var views []dto.BranchStockView
	for _, s := range stocks {
		batch := batchMap[s.BatchId]
		views = append(views, dto.BranchStockView{
			StockId:          s.StockId,
			BatchId:          s.BatchId,
			MedicineId:       s.MedicineId,
			BranchId:         s.BranchId,
			Quantity:         s.Quantity,
			ReservedQuantity: s.ReservedQuantity,
			BatchNumber:      batch.BatchNumber,
			ExpiryDate:       batch.ExpiryDate,
			SellingPrice:     batch.SellingPrice,
			BuyingPrice:      batch.BuyingPrice,
			BatchStatus:      batch.Status,
		})
	}

	// 4. FEFO sort — earliest expiry first
	for i := 0; i < len(views)-1; i++ {
		for j := 0; j < len(views)-i-1; j++ {
			if views[j].ExpiryDate.After(views[j+1].ExpiryDate) {
				views[j], views[j+1] = views[j+1], views[j]
			}
		}
	}

	return views, nil
}

// DB_GetActiveStockByMedicineID returns total available qty across all stock for a medicine in a branch.
func DB_GetActiveStockByMedicineID(medicineID, branchId string) (int, error) {
	views, err := DB_GetAvailableBatchesFEFO(medicineID, branchId)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, v := range views {
		total += v.Quantity - v.ReservedQuantity
	}
	return total, nil
}

// DB_DeductFromBatchAtomic atomically deducts from a BranchStock record by its batchId and branchId.
// It decrements BOTH quantity and reservedQuantity (reservation finalization).
func DB_DeductFromBatchAtomic(batchId string, branchId string, deductAmount int) (int, error) {
	filter := bson.M{
		"batchId":  batchId,
		"branchId": branchId,
		"quantity": bson.M{"$gte": deductAmount},
	}
	update := bson.M{
		"$inc": bson.M{
			"quantity":         -deductAmount,
			"reservedQuantity": -deductAmount,
		},
		"$set": bson.M{"updatedAt": time.Now()},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated dto.BranchStock
	err := dbConfigs.BranchStockCollection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&updated)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, fmt.Errorf("insufficient stock in branch stock record")
		}
		return 0, err
	}
	return updated.Quantity, nil
}

// DB_ReserveStockFEFO atomically reserves stock using FEFO across BranchStock records.
// Increments reservedQuantity so another concurrent bill cannot claim the same pills.
func DB_ReserveStockFEFO(medicineID, branchId string, requiredQty int) ([]dto.BillItem, error) {
	var billItems []dto.BillItem
	remainingToReserve := requiredQty

	for remainingToReserve > 0 {
		views, err := DB_GetAvailableBatchesFEFO(medicineID, branchId)
		if err != nil {
			return nil, err
		}
		if len(views) == 0 {
			if len(billItems) > 0 {
				DB_RevertStockReservation(billItems)
			}
			return nil, fmt.Errorf("insufficient stock for medicine %s in branch %s", medicineID, branchId)
		}

		reservedInThisPass := false
		for _, v := range views {
			if remainingToReserve <= 0 {
				break
			}
			available := v.Quantity - v.ReservedQuantity
			if available <= 0 {
				continue
			}
			reserveAmount := available
			if remainingToReserve < available {
				reserveAmount = remainingToReserve
			}
			// Atomically lock on BranchStock by stockId
			filter := bson.M{
				"stockId": v.StockId,
				"$expr": bson.M{
					"$gte": bson.A{
						bson.M{"$subtract": bson.A{"$quantity", "$reservedQuantity"}},
						reserveAmount,
					},
				},
			}
			update := bson.M{
				"$inc": bson.M{"reservedQuantity": reserveAmount},
				"$set": bson.M{"updatedAt": time.Now()},
			}
			res, err := dbConfigs.BranchStockCollection.UpdateOne(context.Background(), filter, update)
			if err != nil || res.ModifiedCount == 0 {
				continue // concurrency collision — try next view
			}
			billItems = append(billItems, dto.BillItem{
				MedicineID: v.MedicineId,
				BatchID:    v.BatchId,
				StockID:    v.StockId,
				Quantity:   reserveAmount,
				Price:      v.SellingPrice,
			})
			remainingToReserve -= reserveAmount
			reservedInThisPass = true
			if remainingToReserve <= 0 {
				break
			}
		}
		if !reservedInThisPass && remainingToReserve > 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}
	return billItems, nil
}

// DB_RevertStockReservation releases reservations on BranchStock records.
// Uses StockID (preferred) or falls back to a branchId+batchId lookup.
func DB_RevertStockReservation(items []dto.BillItem) {
	for _, item := range items {
		if item.StockID != "" {
			filter := bson.M{"stockId": item.StockID, "reservedQuantity": bson.M{"$gte": item.Quantity}}
			update := bson.M{
				"$inc": bson.M{"reservedQuantity": -item.Quantity},
				"$set": bson.M{"updatedAt": time.Now()},
			}
			_ = dbConfigs.BranchStockCollection.FindOneAndUpdate(context.Background(), filter, update)
		}
	}
}

// DB_DeductStockFEFO is used by the standalone /billing/deduct endpoint.
// For the primary billing flow, use DB_ReserveStockFEFO + ConfirmBill instead.
func DB_DeductStockFEFO(medicineID string, requiredQty int, billId string, branchId string) ([]dto.BillItem, error) {
	var billItems []dto.BillItem
	remainingToDeduct := requiredQty

	for remainingToDeduct > 0 {
		views, err := DB_GetAvailableBatchesFEFO(medicineID, branchId)
		if err != nil {
			return nil, err
		}
		if len(views) == 0 {
			if len(billItems) > 0 {
				return billItems, fmt.Errorf("insufficient stock: partially deducted, no more batches")
			}
			return nil, fmt.Errorf("insufficient stock in branch %s", branchId)
		}

		totalAvailable := 0
		for _, v := range views {
			totalAvailable += v.Quantity - v.ReservedQuantity
		}
		if totalAvailable < remainingToDeduct {
			return nil, fmt.Errorf("insufficient stock: required %d, available %d", remainingToDeduct, totalAvailable)
		}

		deductedInThisPass := false
		for _, v := range views {
			if remainingToDeduct <= 0 {
				break
			}
			
			available := v.Quantity - v.ReservedQuantity
			if available <= 0 {
				continue
			}
			
			deductAmount := available
			if deductAmount > remainingToDeduct {
				deductAmount = remainingToDeduct
			}
			
			_, err = DB_DeductFromBatchAtomic(v.BatchId, branchId, deductAmount)
			if err != nil {
				continue
			}
			billItems = append(billItems, dto.BillItem{
				MedicineID: v.MedicineId,
				BatchID:    v.BatchId,
				StockID:    v.StockId,
				Quantity:   deductAmount,
				Price:      v.SellingPrice,
			})
			// Write SALE movement
			ctx := context.Background()
			movementId, mErr := GenerateId(ctx, "stock_movements", "MOV")
			if mErr == nil {
				_ = DB_CreateStockMovement(dto.StockMovementModel{
					ID: primitive.NewObjectID(), MovementId: movementId,
					BatchId: v.BatchId, MedicineId: v.MedicineId, BranchId: branchId,
					Type: dto.MovementSale, Quantity: deductAmount,
					ReferenceId: billId, ReferenceType: "BILL",
					Notes: fmt.Sprintf("FEFO sale — bill %s", billId), CreatedAt: time.Now(),
				})
			}
			remainingToDeduct -= deductAmount
			deductedInThisPass = true
			if remainingToDeduct <= 0 {
				break
			}
		}
		if !deductedInThisPass && remainingToDeduct > 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}
	return billItems, nil
}

// DB_CheckStockAndCalculatePrice previews what FEFO would allocate — no writes.
func DB_CheckStockAndCalculatePrice(medicineID, branchId string, requiredQty int) ([]dto.BillItem, error) {
	views, err := DB_GetAvailableBatchesFEFO(medicineID, branchId)
	if err != nil {
		return nil, err
	}
	totalAvailable := 0
	for _, v := range views {
		totalAvailable += v.Quantity - v.ReservedQuantity
	}
	if totalAvailable < requiredQty {
		return nil, fmt.Errorf("insufficient stock: required %d, available %d", requiredQty, totalAvailable)
	}
	var billItems []dto.BillItem
	remainingToDeduct := requiredQty
	for _, v := range views {
		if remainingToDeduct <= 0 {
			break
		}
		available := v.Quantity - v.ReservedQuantity
		deductFromBatch := available
		if remainingToDeduct < available {
			deductFromBatch = remainingToDeduct
		}
		billItems = append(billItems, dto.BillItem{
			MedicineID: v.MedicineId,
			BatchID:    v.BatchId,
			StockID:    v.StockId,
			Quantity:   deductFromBatch,
			Price:      v.SellingPrice,
		})
		remainingToDeduct -= deductFromBatch
	}
	return billItems, nil
}


func DB_CreateBill(bill dto.BillModel) error {
	_, err := dbConfigs.BillCollection.InsertOne(context.Background(), bill)
	return err
}

func DB_GetBillByBillId(billId string, branchId string) (*dto.BillModel, error) {
	var bill dto.BillModel
	filter := bson.M{"billId": billId}
	if branchId != "" {
		filter["branchId"] = branchId
	}
	err := dbConfigs.BillCollection.FindOne(context.Background(), filter).Decode(&bill)
	if err != nil {
		return nil, err
	}
	return &bill, nil
}

func DB_UpdateBillStatus(billId string, status string, branchId string) error {
	filter := bson.M{"billId": billId}
	if branchId != "" {
		filter["branchId"] = branchId
	}
	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}
	_, err := dbConfigs.BillCollection.UpdateOne(context.Background(), filter, update)
	return err
}

// DB_RevertStockDeduction re-adds quantities to BranchStock for a failed/cancelled bill.
func DB_RevertStockDeduction(billItems []dto.BillItem) error {
	for _, item := range billItems {
		if item.StockID != "" {
			filter := bson.M{"stockId": item.StockID}
			update := bson.M{
				"$inc": bson.M{"quantity": item.Quantity},
				"$set": bson.M{"updatedAt": time.Now()},
			}
			_, _ = dbConfigs.BranchStockCollection.UpdateOne(context.Background(), filter, update)
		}
	}
	return nil
}

// DB_GetMedicineNamesByIDs returns a map of medicineId -> medicine name for the given IDs.
func DB_GetMedicineNamesByIDs(medicineIDs []string) (map[string]string, error) {
	ctx := context.Background()
	result := make(map[string]string)
	if len(medicineIDs) == 0 {
		return result, nil
	}

	// Build an $in filter using the custom string field "medicineid"
	filter := bson.M{"medicineid": bson.M{"$in": medicineIDs}}
	cursor, err := dbConfigs.MedicineCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var medicines []dto.MedicineModel
	if err = cursor.All(ctx, &medicines); err != nil {
		return nil, err
	}

	for _, m := range medicines {
		result[m.MedicineId] = m.Name
	}
	return result, nil
}

// DB_WriteSaleMovement writes a single SALE StockMovement for a confirmed bill item.
// This is called from ConfirmBill after each atomic deduction succeeds.
func DB_WriteSaleMovement(item dto.BillItem, billId string, branchId string) error {
	ctx := context.Background()
	movementId, err := GenerateId(ctx, "stock_movements", "MOV")
	if err != nil {
		return err
	}
	return DB_CreateStockMovement(dto.StockMovementModel{
		ID:            primitive.NewObjectID(),
		MovementId:    movementId,
		BatchId:       item.BatchID,
		MedicineId:    item.MedicineID,
		BranchId:      branchId,
		Type:          dto.MovementSale,
		Quantity:      item.Quantity,
		ReferenceId:   billId,
		ReferenceType: "BILL",
		Notes:         fmt.Sprintf("Confirmed sale — bill %s", billId),
		CreatedAt:     time.Now(),
	})
}
