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
	
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "medicine_batches",
			"localField":   "medicineid",
			"foreignField": "medicineId",
			"as":           "batches",
		}}},
		bson.D{{Key: "$addFields", Value: bson.M{
			"totalStock": bson.M{
				"$reduce": bson.M{
					"input":        "$batches",
					"initialValue": 0,
					"in": bson.M{
						"$add": []interface{}{
							"$$value",
							bson.M{
								"$cond": []interface{}{
									bson.M{"$eq": []interface{}{"$$this.status", "ACTIVE"}},
									"$$this.quantity",
									0,
								},
							},
						},
					},
				},
			},
		}}},
		bson.D{{Key: "$match", Value: bson.M{
			"$expr": bson.M{
				"$lte": []interface{}{"$totalStock", "$minStockLevel"},
			},
		}}},
	}

	cursor, err := dbConfigs.MedicineCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var medicines []dto.MedicineModel
	if err = cursor.All(ctx, &medicines); err != nil {
		return nil, err
	}
	
	return medicines, nil
}

func DB_CreateMedicineBatch(batch dto.MedicineBatchModel) error {
	_, err := dbConfigs.MedicineBatchCollection.InsertOne(context.Background(), batch)
	return err
}

func DB_GetBatchesByMedicineID(medicineID string) ([]dto.MedicineBatchModel, error) {
	ctx := context.Background()
	filter := bson.M{"medicineId": medicineID}
	
	cursor, err := dbConfigs.MedicineBatchCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var batches []dto.MedicineBatchModel
	if err = cursor.All(ctx, &batches); err != nil {
		return nil, err
	}
	return batches, nil
}

func DB_GetAvailableBatchesFEFO(medicineID string) ([]dto.MedicineBatchModel, error) {
	ctx := context.Background()
	filter := bson.M{
		"medicineId": medicineID,
		"quantity":   bson.M{"$gt": 0},
		"expiryDate": bson.M{"$gt": time.Now()},
		"status":     "ACTIVE",
	}
	
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "expiryDate", Value: 1}}) // Ascending order (First Expiring First)
	
	cursor, err := dbConfigs.MedicineBatchCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var batches []dto.MedicineBatchModel
	if err = cursor.All(ctx, &batches); err != nil {
		return nil, err
	}
	return batches, nil
}

func DB_GetActiveStockByMedicineID(medicineID string) (int, error) {
	ctx := context.Background()
	filter := bson.M{
		"medicineId": medicineID,
		"status":     "ACTIVE",
	}
	
	cursor, err := dbConfigs.MedicineBatchCollection.Find(ctx, filter)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)
	
	var batches []dto.MedicineBatchModel
	if err = cursor.All(ctx, &batches); err != nil {
		return 0, err
	}
	
	totalStock := 0
	for _, batch := range batches {
		totalStock += batch.Quantity
	}
	
	return totalStock, nil
}

func DB_DeductFromBatchAtomic(batchID primitive.ObjectID, deductAmount int) (int, error) {
	filter := bson.M{
		"_id":      batchID,
		"quantity": bson.M{"$gte": deductAmount},
	}
	update := bson.M{
		"$inc": bson.M{"quantity": -deductAmount},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedBatch dto.MedicineBatchModel
	err := dbConfigs.MedicineBatchCollection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&updatedBatch)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, fmt.Errorf("insufficient stock or batch not found")
		}
		return 0, err
	}

	// If quantity is now 0, update status
	if updatedBatch.Quantity == 0 {
		_, _ = dbConfigs.MedicineBatchCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": batchID},
			bson.M{"$set": bson.M{"status": "OUT_OF_STOCK"}},
		)
	}

	return updatedBatch.Quantity, nil
}

// DB_DeductStockFEFO deducts stock from the earliest-expiring batches (FEFO order).
// It writes a SALE StockMovement for each batch deducted.
// billId and branchId are embedded in the movement as the reference.
func DB_DeductStockFEFO(medicineID string, requiredQty int, billId string, branchId string) ([]dto.BillItem, error) {
	var billItems []dto.BillItem
	remainingToDeduct := requiredQty

	// Use a loop to handle potential concurrency issues by re-fetching if an atomic update fails
	for remainingToDeduct > 0 {
		batches, err := DB_GetAvailableBatchesFEFO(medicineID)
		if err != nil {
			return nil, err
		}

		if len(batches) == 0 {
			if len(billItems) > 0 {
				return billItems, fmt.Errorf("insufficient stock: partially deducted %d, but no more available batches for remaining %d", requiredQty-remainingToDeduct, remainingToDeduct)
			}
			return nil, fmt.Errorf("insufficient stock: no available batches found")
		}

		// Calculate total available in current snapshot
		totalAvailable := 0
		for _, b := range batches {
			totalAvailable += b.Quantity
		}

		if totalAvailable < remainingToDeduct {
			if len(billItems) > 0 {
				return billItems, fmt.Errorf("insufficient stock: partially deducted, need %d more, but only %d available", remainingToDeduct, totalAvailable)
			}
			return nil, fmt.Errorf("insufficient stock: required %d, available %d", remainingToDeduct, totalAvailable)
		}

		deductedInThisPass := false
		for _, b := range batches {
			if remainingToDeduct <= 0 {
				break
			}

			deductFromBatch := b.Quantity
			if remainingToDeduct < b.Quantity {
				deductFromBatch = remainingToDeduct
			}

			// Atomic deduction
			_, err := DB_DeductFromBatchAtomic(b.ID, deductFromBatch)
			if err != nil {
				// If this batch failed (e.g., someone else took stock), move to the next batch
				continue
			}

			billItems = append(billItems, dto.BillItem{
				MedicineID: b.MedicineID,
				BatchID:    b.ID.Hex(),
				Quantity:   deductFromBatch,
				Price:      b.SellingPrice,
			})

			// ── Write SALE movement to audit ledger ──
			ctx := context.Background()
			movementId, mErr := GenerateId(ctx, "stock_movements", "MOV")
			if mErr == nil {
				_ = DB_CreateStockMovement(dto.StockMovementModel{
					ID:            primitive.NewObjectID(),
					MovementId:    movementId,
					BatchId:       b.ID.Hex(),
					MedicineId:    b.MedicineID,
					BranchId:      branchId,
					Type:          dto.MovementSale,
					Quantity:      deductFromBatch,
					ReferenceId:   billId,
					ReferenceType: "BILL",
					Notes:         fmt.Sprintf("FEFO sale — bill %s", billId),
					CreatedAt:     time.Now(),
				})
			}

			remainingToDeduct -= deductFromBatch
			deductedInThisPass = true

			// If we fully satisfied the request, we can break
			if remainingToDeduct <= 0 {
				break
			}
		}

		// If we went through all batches but couldn't deduct anything (all failed due to contention),
		// we should avoid an infinite loop. We'll re-fetch anyway in the next outer loop iteration.
		if !deductedInThisPass && remainingToDeduct > 0 {
			time.Sleep(10 * time.Millisecond) // Tiny backoff
		}
	}

	return billItems, nil
}

func DB_CheckStockAndCalculatePrice(medicineID string, requiredQty int) ([]dto.BillItem, error) {
batches, err := DB_GetAvailableBatchesFEFO(medicineID)
if err != nil {
return nil, err
}

totalAvailable := 0
for _, b := range batches {
totalAvailable += b.Quantity
}

if totalAvailable < requiredQty {
return nil, fmt.Errorf("insufficient stock: required %d, available %d", requiredQty, totalAvailable)
}

var billItems []dto.BillItem
remainingToDeduct := requiredQty

for _, b := range batches {
if remainingToDeduct <= 0 {
break
}

deductFromBatch := b.Quantity
if remainingToDeduct < b.Quantity {
deductFromBatch = remainingToDeduct
}

billItems = append(billItems, dto.BillItem{
MedicineID: b.MedicineID,
BatchID:    b.ID.Hex(),
Quantity:   deductFromBatch,
Price:      b.SellingPrice,
})

remainingToDeduct -= deductFromBatch
}

return billItems, nil
}

func DB_CreateBill(bill dto.BillModel) error {
	_, err := dbConfigs.BillCollection.InsertOne(context.Background(), bill)
	return err
}

func DB_GetBillByBillId(billId string) (*dto.BillModel, error) {
	var bill dto.BillModel
	err := dbConfigs.BillCollection.FindOne(context.Background(), bson.M{"billId": billId}).Decode(&bill)
	if err != nil {
	return nil, err
	}
	return &bill, nil
}

func DB_UpdateBillStatus(billId string, status string) error {
	filter := bson.M{"billId": billId}
	update := bson.M{
	"$set": bson.M{
	"status":    status,
	"updatedAt": time.Now(),
	},
}
	_, err := dbConfigs.BillCollection.UpdateOne(context.Background(), filter, update)
	return err
}

func DB_RevertStockDeduction(billItems []dto.BillItem) error {
	for _, item := range billItems {
	batchObjID, err := primitive.ObjectIDFromHex(item.BatchID)
	if err != nil {
	return err
}

	filter := bson.M{"_id": batchObjID}
	update := bson.M{
	"$inc": bson.M{"quantity": item.Quantity},
	"$set": bson.M{"status": "ACTIVE"}, // Ensure status is active in case it was out of stock
}

	_, err = dbConfigs.MedicineBatchCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
		return err
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
