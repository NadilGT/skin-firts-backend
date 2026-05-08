package dbConfigs

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ensureSupplierMedicinePriceIndexes creates the required indexes on the
// supplier_medicine_prices collection at application start-up.
//
// Indexes:
//  1. { supplierId: 1 }           — fast look-up by supplier
//  2. { medicineId: 1 }           — fast look-up by medicine
//  3. { supplierId: 1, medicineId: 1 } unique — prevents duplicate pricing rows
func ensureSupplierMedicinePriceIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "supplierId", Value: 1}},
			Options: options.Index().SetName("idx_smp_supplierId"),
		},
		{
			Keys:    bson.D{{Key: "medicineId", Value: 1}},
			Options: options.Index().SetName("idx_smp_medicineId"),
		},
		{
			Keys: bson.D{
				{Key: "supplierId", Value: 1},
				{Key: "medicineId", Value: 1},
			},
			Options: options.Index().
				SetName("idx_smp_supplier_medicine_unique").
				SetUnique(true),
		},
	}

	_, err := SupplierMedicinePriceCollection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("[DB INDEX] ⚠️  Failed to create supplier_medicine_prices indexes: %v", err)
		return
	}
	log.Println("[DB INDEX] ✅ supplier_medicine_prices indexes ensured")
}
