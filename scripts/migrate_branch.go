// scripts/migrate_branch.go
// One-time migration: tags all legacy documents with branchId = "BRN-001" (Main Branch).
// Run once: go run ./scripts/migrate_branch.go
// Safe to re-run — only updates documents where branchId does not exist.

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const mainBranchId = "BRN-001"

func main() {
	_ = godotenv.Load(".env")

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("❌ MONGODB_URI not set in environment")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("test") // change if your DB name differs

	// Collections to migrate — all documents missing branchId get tagged BRN-001
	collections := []string{
		"bills",
		"medicine_batches",
		"appointments",
		"doctor_info",
		"purchase_orders",
		"grn",
		"suppliers",
		"stock_transfers",
	}

	filter := bson.M{"branchId": bson.M{"$exists": false}}
	update := bson.M{"$set": bson.M{"branchId": mainBranchId}}

	totalUpdated := int64(0)
	for _, colName := range collections {
		col := db.Collection(colName)
		res, err := col.UpdateMany(ctx, filter, update)
		if err != nil {
			fmt.Printf("⚠️  [%s] Error: %v\n", colName, err)
			continue
		}
		fmt.Printf("✅ [%-20s] Tagged %d documents with branchId=%s\n", colName, res.ModifiedCount, mainBranchId)
		totalUpdated += res.ModifiedCount
	}

	fmt.Printf("\n🎉 Migration complete. Total documents updated: %d\n", totalUpdated)
}
