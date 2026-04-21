package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ──────────────────────────────────────────────
//  Top Selling Medicines
// ──────────────────────────────────────────────

func DB_GetTopSellingMedicines(query dto.AnalyticsQuery) ([]dto.TopSellingItem, error) {
	ctx := context.Background()

	matchFilter := bson.M{"status": "CONFIRMED"}
	if query.BranchId != "" {
		matchFilter["branchId"] = query.BranchId
	}
	applyDateRange(matchFilter, query.From, query.To)

	limit := query.Limit
	if limit <= 0 {
		limit = 10
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchFilter}},
		bson.D{{Key: "$unwind", Value: "$items"}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":          "$items.medicineId",
			"totalQtySold": bson.M{"$sum": "$items.quantity"},
			"totalRevenue": bson.M{"$sum": bson.M{"$multiply": []interface{}{"$items.quantity", "$items.price"}}},
		}}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "medicines",
			"localField":   "_id",
			"foreignField": "medicineid",
			"as":           "medicine",
		}}},
		bson.D{{Key: "$unwind", Value: bson.M{"path": "$medicine", "preserveNullAndEmptyArrays": true}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "totalQtySold", Value: -1}}}},
		bson.D{{Key: "$limit", Value: int64(limit)}},
		bson.D{{Key: "$project", Value: bson.M{
			"medicineId":   "$_id",
			"medicineName": "$medicine.name",
			"totalQtySold": 1,
			"totalRevenue": 1,
		}}},
	}

	cursor, err := dbConfigs.BillCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []dto.TopSellingItem
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// ──────────────────────────────────────────────
//  Sales Report (daily / monthly)
// ──────────────────────────────────────────────

func DB_GetSalesReport(query dto.AnalyticsQuery) ([]dto.SalesReportItem, error) {
	ctx := context.Background()

	matchFilter := bson.M{"status": "CONFIRMED"}
	if query.BranchId != "" {
		matchFilter["branchId"] = query.BranchId
	}
	applyDateRange(matchFilter, query.From, query.To)

	format := "%Y-%m-%d" // daily
	if query.Period == "monthly" {
		format = "%Y-%m"
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchFilter}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{"format": format, "date": "$createdAt"},
			},
			"totalBills":    bson.M{"$sum": 1},
			"totalRevenue":  bson.M{"$sum": "$netTotal"},
			"totalDiscount": bson.M{"$sum": "$discount"},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
		bson.D{{Key: "$project", Value: bson.M{
			"period":        "$_id",
			"totalBills":    1,
			"totalRevenue":  1,
			"totalDiscount": 1,
		}}},
	}

	cursor, err := dbConfigs.BillCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []dto.SalesReportItem
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// ──────────────────────────────────────────────
//  Profit Margin Report
// ──────────────────────────────────────────────

func DB_GetProfitMarginReport(query dto.AnalyticsQuery) ([]dto.ProfitMarginItem, error) {
	ctx := context.Background()

	matchFilter := bson.M{"status": "CONFIRMED"}
	if query.BranchId != "" {
		matchFilter["branchId"] = query.BranchId
	}
	applyDateRange(matchFilter, query.From, query.To)

	// Join bill items with batch buying prices
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchFilter}},
		bson.D{{Key: "$unwind", Value: "$items"}},
		// lookup the batch to get buying price
		bson.D{{Key: "$lookup", Value: bson.M{
			"from": "medicine_batches",
			"let":  bson.M{"batchId": "$items.batchId"},
			"pipeline": mongo.Pipeline{
				bson.D{{Key: "$match", Value: bson.M{"$expr": bson.M{"$eq": []interface{}{bson.M{"$toString": "$_id"}, "$$batchId"}}}}},
				bson.D{{Key: "$project", Value: bson.M{"buyingPrice": 1}}},
			},
			"as": "batchInfo",
		}}},
		bson.D{{Key: "$unwind", Value: bson.M{"path": "$batchInfo", "preserveNullAndEmptyArrays": true}}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":          "$items.medicineId",
			"totalQtySold": bson.M{"$sum": "$items.quantity"},
			"totalRevenue": bson.M{"$sum": bson.M{"$multiply": []interface{}{"$items.quantity", "$items.price"}}},
			"totalCost":    bson.M{"$sum": bson.M{"$multiply": []interface{}{"$items.quantity", "$batchInfo.buyingPrice"}}},
		}}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "medicines",
			"localField":   "_id",
			"foreignField": "medicineid",
			"as":           "medicine",
		}}},
		bson.D{{Key: "$unwind", Value: bson.M{"path": "$medicine", "preserveNullAndEmptyArrays": true}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "totalRevenue", Value: -1}}}},
		bson.D{{Key: "$project", Value: bson.M{
			"medicineId":    "$_id",
			"medicineName":  "$medicine.name",
			"totalQtySold":  1,
			"totalRevenue":  1,
			"totalCost":     1,
			"grossProfit":   bson.M{"$subtract": []interface{}{"$totalRevenue", "$totalCost"}},
			"profitMarginPct": bson.M{
				"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$totalRevenue", 0}},
					0,
					bson.M{"$multiply": []interface{}{
						bson.M{"$divide": []interface{}{
							bson.M{"$subtract": []interface{}{"$totalRevenue", "$totalCost"}},
							"$totalRevenue",
						}},
						100,
					}},
				},
			},
		}}},
	}

	cursor, err := dbConfigs.BillCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []dto.ProfitMarginItem
	if err = cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// ──────────────────────────────────────────────
//  Helper
// ──────────────────────────────────────────────

func applyDateRange(filter bson.M, from, to string) {
	if from == "" && to == "" {
		return
	}
	dateFilter := bson.M{}
	if from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			dateFilter["$gte"] = t
		}
	}
	if to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			dateFilter["$lte"] = t.Add(24 * time.Hour)
		}
	}
	if len(dateFilter) > 0 {
		filter["createdAt"] = dateFilter
	}
}
