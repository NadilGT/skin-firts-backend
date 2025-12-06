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
			"price":                medicine.Price,
			"stockQuantity":        medicine.StockQuantity,
			"minStockLevel":        medicine.MinStockLevel,
			"expiryDate":           medicine.ExpiryDate,
			"batchNumber":          medicine.BatchNumber,
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
	
	// Find medicines where stockQuantity <= minStockLevel
	filter := bson.M{
		"$expr": bson.M{
			"$lte": []interface{}{"$stockQuantity", "$minStockLevel"},
		},
	}
	
	cursor, err := dbConfigs.MedicineCollection.Find(ctx, filter)
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