package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TotalRevenueResult holds the combined revenue from pharmacy bills and hospital bills.
type TotalRevenueResult struct {
	Date              string  `json:"date"`
	BranchId          string  `json:"branchId"`
	PharmacyRevenue   float64 `json:"pharmacyRevenue"`   // sum of grandTotal from bills (paymentStatus=PAID)
	HospitalRevenue   float64 `json:"hospitalRevenue"`   // sum of totalAmount from hospital_bills (confirm=true)
	TotalRevenue      float64 `json:"totalRevenue"`      // pharmacyRevenue + hospitalRevenue
	PharmacyBillCount int64   `json:"pharmacyBillCount"` // number of qualifying pharmacy bills
	HospitalBillCount int64   `json:"hospitalBillCount"` // number of qualifying hospital bills
}

// DB_GetTotalRevenue aggregates revenue from both billing collections for a specific date and branch.
// date format: "YYYY-MM-DD". branchId is required.
func DB_GetTotalRevenue(branchId, date string) (*TotalRevenueResult, error) {
	ctx := context.Background()

	// Parse date into day start/end bounds (UTC-aware range)
	var start, end time.Time
	if date == "" {
		y, m, d := time.Now().Date()
		start = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	} else {
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			return nil, err
		}
		start = t.UTC()
	}
	end = start.Add(24 * time.Hour)

	result := &TotalRevenueResult{
		Date:     start.Format("2006-01-02"),
		BranchId: branchId,
	}

	// ── 1. Pharmacy Bills: paymentStatus = "PAID" ──────────────────────────
	billFilter := bson.M{
		"paymentStatus": "PAID",
		"createdAt":     bson.M{"$gte": start, "$lt": end},
	}
	if branchId != "" {
		billFilter["branchId"] = branchId
	}

	billPipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: billFilter}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":      nil,
			"revenue":  bson.M{"$sum": "$grandTotal"},
			"count":    bson.M{"$sum": 1},
		}}},
	}

	billCursor, err := dbConfigs.BillCollection.Aggregate(ctx, billPipeline)
	if err != nil {
		return nil, err
	}
	defer billCursor.Close(ctx)

	type aggResult struct {
		Revenue float64 `bson:"revenue"`
		Count   int64   `bson:"count"`
	}
	if billCursor.Next(ctx) {
		var r aggResult
		if err := billCursor.Decode(&r); err != nil {
			return nil, err
		}
		result.PharmacyRevenue = r.Revenue
		result.PharmacyBillCount = r.Count
	}

	// ── 2. Hospital Bills: confirm = true ──────────────────────────────────
	hospitalFilter := bson.M{
		"confirm":   true,
		"createdAt": bson.M{"$gte": start, "$lt": end},
	}
	if branchId != "" {
		hospitalFilter["branchId"] = branchId
	}

	hospitalPipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: hospitalFilter}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":     nil,
			"revenue": bson.M{"$sum": "$totalAmount"},
			"count":   bson.M{"$sum": 1},
		}}},
	}

	hospitalCursor, err := dbConfigs.HospitalBillCollection.Aggregate(ctx, hospitalPipeline)
	if err != nil {
		return nil, err
	}
	defer hospitalCursor.Close(ctx)

	if hospitalCursor.Next(ctx) {
		var r aggResult
		if err := hospitalCursor.Decode(&r); err != nil {
			return nil, err
		}
		result.HospitalRevenue = r.Revenue
		result.HospitalBillCount = r.Count
	}

	result.TotalRevenue = result.PharmacyRevenue + result.HospitalRevenue
	return result, nil
}
