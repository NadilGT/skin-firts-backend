package dao

import (
	"context"
	"errors"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ──────────────────────────────────────────────
//
//	Supplier Medicine Price DAO
//
// ──────────────────────────────────────────────

// DB_CreateSupplierMedicinePrice inserts a new price record.
// Returns a duplicate-key error if the (supplierId, medicineId) pair already exists.
func DB_CreateSupplierMedicinePrice(p dto.SupplierMedicinePrice) error {
	_, err := dbConfigs.SupplierMedicinePriceCollection.InsertOne(context.Background(), p)
	return err
}

// DB_GetSupplierMedicinePrices returns all price records matching the optional filters.
func DB_GetSupplierMedicinePrices(query dto.SearchSupplierMedicinePriceQuery) ([]dto.SupplierMedicinePrice, error) {
	ctx := context.Background()
	filter := bson.M{}

	if query.SupplierId != "" {
		filter["supplierId"] = query.SupplierId
	}
	if query.MedicineId != "" {
		filter["medicineId"] = query.MedicineId
	}
	if query.IsActive != nil {
		filter["isActive"] = *query.IsActive
	}

	findOpts := options.Find().SetSort(bson.D{
		{Key: "supplierId", Value: 1},
		{Key: "medicineName", Value: 1},
	})
	cursor, err := dbConfigs.SupplierMedicinePriceCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []dto.SupplierMedicinePrice
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// DB_GetSupplierMedicinePriceByID fetches a single record by its generated priceId (e.g. "SMP-001").
func DB_GetSupplierMedicinePriceByID(priceId string) (*dto.SupplierMedicinePrice, error) {
	if priceId == "" {
		return nil, errors.New("priceId is required")
	}
	var p dto.SupplierMedicinePrice
	err := dbConfigs.SupplierMedicinePriceCollection.
		FindOne(context.Background(), bson.M{"priceId": priceId}).
		Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// DB_GetSupplierMedicinePriceBySupplierAndMedicine is the key lookup used during
// PO creation — returns only active price records.
func DB_GetSupplierMedicinePriceBySupplierAndMedicine(supplierId, medicineId string) (*dto.SupplierMedicinePrice, error) {
	var p dto.SupplierMedicinePrice
	filter := bson.M{
		"supplierId": supplierId,
		"medicineId": medicineId,
		"isActive":   true,
	}
	err := dbConfigs.SupplierMedicinePriceCollection.
		FindOne(context.Background(), filter).
		Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// DB_UpdateSupplierMedicinePrice applies a partial update (unitPrice and/or isActive)
// identified by the generated priceId string.
func DB_UpdateSupplierMedicinePrice(priceId string, req dto.UpdateSupplierMedicinePriceRequest) error {
	if priceId == "" {
		return errors.New("priceId is required")
	}

	setFields := bson.M{"updatedAt": time.Now()}
	if req.UnitPrice > 0 {
		setFields["unitPrice"] = req.UnitPrice
	}
	if req.IsActive != nil {
		setFields["isActive"] = *req.IsActive
	}

	_, err := dbConfigs.SupplierMedicinePriceCollection.UpdateOne(
		context.Background(),
		bson.M{"priceId": priceId},
		bson.M{"$set": setFields},
	)
	return err
}

// DB_DeleteSupplierMedicinePrice performs a soft delete by setting isActive = false.
func DB_DeleteSupplierMedicinePrice(priceId string) error {
	f := false
	return DB_UpdateSupplierMedicinePrice(priceId, dto.UpdateSupplierMedicinePriceRequest{
		IsActive: &f,
	})
}

// IsDuplicateKeyError checks whether a MongoDB write error is a duplicate-key violation.
func IsDuplicateKeyError(err error) bool {
	var we mongo.WriteException
	if errors.As(err, &we) {
		for _, e := range we.WriteErrors {
			if e.Code == 11000 {
				return true
			}
		}
	}
	return false
}
