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

func DB_UpdateBatchQuantity(batchID primitive.ObjectID, quantity int, status string) error {
	filter := bson.M{"_id": batchID}
	update := bson.M{
		"$set": bson.M{
			"quantity": quantity,
			"status":   status,
		},
	}
	_, err := dbConfigs.MedicineBatchCollection.UpdateOne(context.Background(), filter, update)
	return err
}

func DB_DeductStockFEFO(medicineID string, requiredQty int) ([]dto.BillItem, error) {
	batches, err := DB_GetAvailableBatchesFEFO(medicineID)
	if err != nil {
		return nil, err
	}

	// Calculate total available
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

		newQty := b.Quantity - deductFromBatch
		status := b.Status
		if newQty == 0 {
			status = "OUT_OF_STOCK"
		}

		err = DB_UpdateBatchQuantity(b.ID, newQty, status)
		if err != nil {
			return nil, err
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