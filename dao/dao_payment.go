package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ──────────────────────────────────────────────
//  Bill Queries (extended for POS)
// ──────────────────────────────────────────────

func DB_SearchPharmacyBills(query dto.SearchBillQuery) ([]dto.BillModel, int64, error) {
	ctx := context.Background()
	filter := bson.M{}

	if query.BranchId != "" {
		filter["branchId"] = query.BranchId
	}
	if query.PaymentStatus != "" {
		filter["paymentStatus"] = query.PaymentStatus
	}
	if query.Status != "" {
		filter["status"] = query.Status
	}

	// Date range on createdAt
	if query.From != "" || query.To != "" {
		dateFilter := bson.M{}
		if query.From != "" {
			if t, err := time.Parse("2006-01-02", query.From); err == nil {
				dateFilter["$gte"] = t
			}
		}
		if query.To != "" {
			if t, err := time.Parse("2006-01-02", query.To); err == nil {
				dateFilter["$lte"] = t.Add(24 * time.Hour)
			}
		}
		if len(dateFilter) > 0 {
			filter["createdAt"] = dateFilter
		}
	}

	total, err := dbConfigs.BillCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}

	findOpts := options.Find().
		SetSkip(int64((query.Page - 1) * query.Limit)).
		SetLimit(int64(query.Limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := dbConfigs.BillCollection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var bills []dto.BillModel
	if err = cursor.All(ctx, &bills); err != nil {
		return nil, 0, err
	}
	return bills, total, nil
}

func DB_UpdateBillPayment(billId string, branchId string, req dto.UpdateBillPaymentRequest) error {
	bill, err := DB_GetBillByBillId(billId, branchId)
	if err != nil {
		return err
	}

	newPaid := bill.PaidAmount + req.PaidAmount
	newDue := bill.NetTotal - newPaid
	paymentStatus := "PARTIAL"
	if newDue <= 0 {
		paymentStatus = "PAID"
		newDue = 0
	}

	filter := bson.M{"billId": billId}
	if branchId != "" {
		filter["branchId"] = branchId
	}
	update := bson.M{
		"$set": bson.M{
			"paidAmount":    newPaid,
			"dueAmount":     newDue,
			"paymentStatus": paymentStatus,
			"paymentMethod": req.PaymentMethod,
			"notes":         req.Notes,
			"updatedAt":     time.Now(),
		},
	}
	_, err = dbConfigs.BillCollection.UpdateOne(context.Background(), filter, update)
	return err
}

// ──────────────────────────────────────────────
//  Daily Sales Summary
// ──────────────────────────────────────────────

func DB_GetDailySalesSummary(branchId, date string) (*dto.DailySalesSummary, error) {
	ctx := context.Background()

	var start, end time.Time
	if date == "" {
		y, m, d := time.Now().Date()
		start = time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	} else {
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			return nil, err
		}
		start = t
	}
	end = start.Add(24 * time.Hour)

	matchFilter := bson.M{
		"status":    "CONFIRMED",
		"createdAt": bson.M{"$gte": start, "$lt": end},
	}
	if branchId != "" {
		matchFilter["branchId"] = branchId
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchFilter}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":        nil,
			"totalBills": bson.M{"$sum": 1},
			"totalRevenue": bson.M{"$sum": "$netTotal"},
			"cashRevenue": bson.M{"$sum": bson.M{
				"$cond": []interface{}{bson.M{"$eq": []interface{}{"$paymentMethod", "CASH"}}, "$paidAmount", 0},
			}},
			"cardRevenue": bson.M{"$sum": bson.M{
				"$cond": []interface{}{bson.M{"$eq": []interface{}{"$paymentMethod", "CARD"}}, "$paidAmount", 0},
			}},
			"onlineRevenue": bson.M{"$sum": bson.M{
				"$cond": []interface{}{bson.M{"$eq": []interface{}{"$paymentMethod", "ONLINE"}}, "$paidAmount", 0},
			}},
			"totalDiscount": bson.M{"$sum": "$discount"},
			"totalTax":      bson.M{"$sum": "$tax"},
			"totalDue":      bson.M{"$sum": "$dueAmount"},
		}}},
	}

	cursor, err := dbConfigs.BillCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type rawResult struct {
		TotalBills    int     `bson:"totalBills"`
		TotalRevenue  float64 `bson:"totalRevenue"`
		CashRevenue   float64 `bson:"cashRevenue"`
		CardRevenue   float64 `bson:"cardRevenue"`
		OnlineRevenue float64 `bson:"onlineRevenue"`
		TotalDiscount float64 `bson:"totalDiscount"`
		TotalTax      float64 `bson:"totalTax"`
		TotalDue      float64 `bson:"totalDue"`
	}

	summary := &dto.DailySalesSummary{
		Date:     start.Format("2006-01-02"),
		BranchId: branchId,
	}

	if cursor.Next(ctx) {
		var raw rawResult
		if err = cursor.Decode(&raw); err != nil {
			return nil, err
		}
		summary.TotalBills = raw.TotalBills
		summary.TotalRevenue = raw.TotalRevenue
		summary.CashRevenue = raw.CashRevenue
		summary.CardRevenue = raw.CardRevenue
		summary.OnlineRevenue = raw.OnlineRevenue
		summary.TotalDiscount = raw.TotalDiscount
		summary.TotalTax = raw.TotalTax
		summary.TotalDue = raw.TotalDue
	}
	return summary, nil
}

// ──────────────────────────────────────────────
//  Revenue Summary (date range)
// ──────────────────────────────────────────────

func DB_GetRevenueSummary(branchId, from, to string) (*dto.RevenueSummaryResponse, error) {
	ctx := context.Background()

	matchFilter := bson.M{"status": "CONFIRMED"}
	if branchId != "" {
		matchFilter["branchId"] = branchId
	}
	if from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			if matchFilter["createdAt"] == nil {
				matchFilter["createdAt"] = bson.M{}
			}
			matchFilter["createdAt"].(bson.M)["$gte"] = t
		}
	}
	if to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			if matchFilter["createdAt"] == nil {
				matchFilter["createdAt"] = bson.M{}
			}
			matchFilter["createdAt"].(bson.M)["$lte"] = t.Add(24 * time.Hour)
		}
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchFilter}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$createdAt"},
			},
			"totalBills":    bson.M{"$sum": 1},
			"totalRevenue":  bson.M{"$sum": "$netTotal"},
			"totalDiscount": bson.M{"$sum": "$discount"},
			"totalTax":      bson.M{"$sum": "$tax"},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}

	cursor, err := dbConfigs.BillCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type rawItem struct {
		Period        string  `bson:"_id"`
		TotalBills    int     `bson:"totalBills"`
		TotalRevenue  float64 `bson:"totalRevenue"`
		TotalDiscount float64 `bson:"totalDiscount"`
		TotalTax      float64 `bson:"totalTax"`
	}

	var rawItems []rawItem
	if err = cursor.All(ctx, &rawItems); err != nil {
		return nil, err
	}

	var items []dto.RevenueSummaryItem
	var grand float64
	for _, r := range rawItems {
		items = append(items, dto.RevenueSummaryItem{
			Period:        r.Period,
			TotalBills:    r.TotalBills,
			TotalRevenue:  r.TotalRevenue,
			TotalDiscount: r.TotalDiscount,
			TotalTax:      r.TotalTax,
		})
		grand += r.TotalRevenue
	}

	return &dto.RevenueSummaryResponse{
		BranchId:   branchId,
		From:       from,
		To:         to,
		Items:      items,
		GrandTotal: grand,
	}, nil
}

// DB_GetPendingPayments returns bills with paymentStatus = PARTIAL or PENDING (confirmed bills with outstanding balance).
func DB_GetPendingPayments(branchId string, page, limit int) ([]dto.BillModel, int64, error) {
	query := dto.SearchBillQuery{
		BranchId:      branchId,
		Status:        "CONFIRMED",
		PaymentStatus: "PARTIAL",
		Page:          page,
		Limit:         limit,
	}
	partialBills, partialTotal, err := DB_SearchPharmacyBills(query)
	if err != nil {
		return nil, 0, err
	}
	query.PaymentStatus = "PENDING"
	pendingBills, pendingTotal, err := DB_SearchPharmacyBills(query)
	if err != nil {
		return nil, 0, err
	}
	return append(partialBills, pendingBills...), partialTotal + pendingTotal, nil
}
