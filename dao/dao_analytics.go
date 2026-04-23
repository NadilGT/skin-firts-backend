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
		bson.D{{Key: "$sort", Value: bson.D{{Key: "totalQtySold", Value: -1}}}},
		bson.D{{Key: "$limit", Value: int64(limit)}},
	}

	cursor, err := dbConfigs.BillCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type rawItem struct {
		MedicineId   string  `bson:"_id"`
		TotalQtySold int     `bson:"totalQtySold"`
		TotalRevenue float64 `bson:"totalRevenue"`
	}

	var rawItems []rawItem
	if err = cursor.All(ctx, &rawItems); err != nil {
		return nil, err
	}

	var items []dto.TopSellingItem
	if len(rawItems) == 0 {
		return items, nil
	}

	// Fetch medicine names
	medicineIDs := []string{}
	for _, r := range rawItems {
		medicineIDs = append(medicineIDs, r.MedicineId)
	}
	nameMap, _ := DB_GetMedicineNamesByIDs(medicineIDs)

	for _, r := range rawItems {
		items = append(items, dto.TopSellingItem{
			MedicineID:   r.MedicineId,
			MedicineName: nameMap[r.MedicineId],
			TotalQtySold: r.TotalQtySold,
			TotalRevenue: r.TotalRevenue,
		})
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

	// Group by both medicineId and batchId to calculate accurate costs later
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchFilter}},
		bson.D{{Key: "$unwind", Value: "$items"}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"medicineId": "$items.medicineId",
				"batchId":    "$items.batchId",
			},
			"totalQtySold": bson.M{"$sum": "$items.quantity"},
			"totalRevenue": bson.M{"$sum": bson.M{"$multiply": []interface{}{"$items.quantity", "$items.price"}}},
		}}},
	}

	cursor, err := dbConfigs.BillCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type rawItem struct {
		ID struct {
			MedicineId string `bson:"medicineId"`
			BatchId    string `bson:"batchId"`
		} `bson:"_id"`
		TotalQtySold int     `bson:"totalQtySold"`
		TotalRevenue float64 `bson:"totalRevenue"`
	}

	var rawItems []rawItem
	if err = cursor.All(ctx, &rawItems); err != nil {
		return nil, err
	}

	var finalItems []dto.ProfitMarginItem
	if len(rawItems) == 0 {
		return finalItems, nil
	}

	// Extract IDs for Go-level resolution
	medicineIDsMap := make(map[string]bool)
	batchIDsMap := make(map[string]bool)
	for _, r := range rawItems {
		medicineIDsMap[r.ID.MedicineId] = true
		batchIDsMap[r.ID.BatchId] = true
	}

	medicineIDs := []string{}
	for id := range medicineIDsMap {
		medicineIDs = append(medicineIDs, id)
	}
	
	batchIDs := []string{}
	for id := range batchIDsMap {
		batchIDs = append(batchIDs, id)
	}

	// Fetch medicine names
	nameMap, _ := DB_GetMedicineNamesByIDs(medicineIDs)

	// Fetch batch buying prices
	batchPrices := make(map[string]float64)
	if len(batchIDs) > 0 {
		batchCursor, bErr := dbConfigs.MedicineBatchCollection.Find(ctx, bson.M{"batchId": bson.M{"$in": batchIDs}})
		if bErr == nil {
			defer batchCursor.Close(ctx)
			var batches []dto.MedicineBatch
			if batchCursor.All(ctx, &batches) == nil {
				for _, b := range batches {
					batchPrices[b.BatchId] = b.BuyingPrice
				}
			}
		}
	}

	// Aggregate by medicineId in Go
	type aggMedicine struct {
		TotalQtySold int
		TotalRevenue float64
		TotalCost    float64
	}
	aggregated := make(map[string]*aggMedicine)

	for _, r := range rawItems {
		medId := r.ID.MedicineId
		if aggregated[medId] == nil {
			aggregated[medId] = &aggMedicine{}
		}
		buyingPrice := batchPrices[r.ID.BatchId]
		
		aggregated[medId].TotalQtySold += r.TotalQtySold
		aggregated[medId].TotalRevenue += r.TotalRevenue
		aggregated[medId].TotalCost += float64(r.TotalQtySold) * buyingPrice
	}

	// Format final response
	for medId, agg := range aggregated {
		grossProfit := agg.TotalRevenue - agg.TotalCost
		var marginPct float64
		if agg.TotalRevenue > 0 {
			marginPct = (grossProfit / agg.TotalRevenue) * 100
		}

		finalItems = append(finalItems, dto.ProfitMarginItem{
			MedicineID:      medId,
			MedicineName:    nameMap[medId],
			TotalQtySold:    agg.TotalQtySold,
			TotalRevenue:    agg.TotalRevenue,
			TotalCost:       agg.TotalCost,
			GrossProfit:     grossProfit,
			ProfitMarginPct: marginPct,
		})
	}

	// Sort by revenue descending (basic bubble sort since slice is likely small)
	for i := 0; i < len(finalItems)-1; i++ {
		for j := 0; j < len(finalItems)-i-1; j++ {
			if finalItems[j].TotalRevenue < finalItems[j+1].TotalRevenue {
				finalItems[j], finalItems[j+1] = finalItems[j+1], finalItems[j]
			}
		}
	}

	return finalItems, nil
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
