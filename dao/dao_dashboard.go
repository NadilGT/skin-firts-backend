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
//  Helpers
// ──────────────────────────────────────────────

// buildDayRange returns [startUTC, endUTC) covering `days` calendar days ending
// at the very end of today (UTC).  e.g. days=7 → last 7 full days including today.
func buildDayRange(days int) (start, end time.Time) {
	now := time.Now().UTC()
	// end = start of tomorrow (exclusive upper bound)
	end = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)
	start = end.Add(-time.Duration(days) * 24 * time.Hour)
	return
}

// generateDateSequence returns a slice of "YYYY-MM-DD" strings for every
// calendar day in [start, end).
func generateDateSequence(start, end time.Time) []string {
	var dates []string
	for d := start; d.Before(end); d = d.Add(24 * time.Hour) {
		dates = append(dates, d.Format("2006-01-02"))
	}
	return dates
}

// ──────────────────────────────────────────────
//  API 1 — Appointments time-series
// ──────────────────────────────────────────────

// DB_GetAppointmentsTimeSeries returns last `days` days of appointment counts
// grouped by date (UTC), zero-filled for missing days.
// branchId is required.
func DB_GetAppointmentsTimeSeries(branchId string, days int) ([]dto.AppointmentDataPoint, error) {
	ctx := context.Background()

	start, end := buildDayRange(days)

	matchFilter := bson.M{
		"branchId": branchId,
		"appointmentDate": bson.M{
			"$gte": start,
			"$lt":  end,
		},
	}

	// Aggregate: group by YYYY-MM-DD of appointmentDate, count docs, sort asc.
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchFilter}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{
					"format":   "%Y-%m-%d",
					"date":     "$appointmentDate",
					"timezone": "UTC",
				},
			},
			"count": bson.M{"$sum": 1},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}

	cursor, err := dbConfigs.AppointmentCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type rawPoint struct {
		Date  string `bson:"_id"`
		Count int    `bson:"count"`
	}

	var rawPoints []rawPoint
	if err = cursor.All(ctx, &rawPoints); err != nil {
		return nil, err
	}

	// Build a lookup map for fast zero-fill
	countMap := make(map[string]int, len(rawPoints))
	for _, rp := range rawPoints {
		countMap[rp.Date] = rp.Count
	}

	// Zero-fill: iterate every day in the window
	dates := generateDateSequence(start, end)
	result := make([]dto.AppointmentDataPoint, 0, len(dates))
	for _, d := range dates {
		result = append(result, dto.AppointmentDataPoint{
			Date:  d,
			Count: countMap[d], // 0 when absent
		})
	}
	return result, nil
}

// ──────────────────────────────────────────────
//  API 2 — Revenue time-series
// ──────────────────────────────────────────────

// DB_GetRevenueTimeSeries returns last `days` days of total revenue (pharmacy +
// hospital bills) grouped by date (UTC), zero-filled for missing days.
// Only completed payments are counted:
//   - Pharmacy bills: paymentStatus = "PAID"
//   - Hospital bills: confirm = true
//
// branchId is required.
func DB_GetRevenueTimeSeries(branchId string, days int) ([]dto.RevenueDataPoint, error) {
	ctx := context.Background()

	start, end := buildDayRange(days)

	// ── 1. Pharmacy bills ──────────────────────────────────────────────────
	pharmacyFilter := bson.M{
		"branchId":      branchId,
		"paymentStatus": "PAID",
		"createdAt":     bson.M{"$gte": start, "$lt": end},
	}

	pharmPipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: pharmacyFilter}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{
					"format":   "%Y-%m-%d",
					"date":     "$createdAt",
					"timezone": "UTC",
				},
			},
			"revenue": bson.M{"$sum": "$grandTotal"},
		}}},
	}

	pharmCursor, err := dbConfigs.BillCollection.Aggregate(ctx, pharmPipeline)
	if err != nil {
		return nil, err
	}
	defer pharmCursor.Close(ctx)

	type rawRev struct {
		Date    string  `bson:"_id"`
		Revenue float64 `bson:"revenue"`
	}

	revenueMap := make(map[string]float64)
	var pharmRaw []rawRev
	if err = pharmCursor.All(ctx, &pharmRaw); err != nil {
		return nil, err
	}
	for _, r := range pharmRaw {
		revenueMap[r.Date] += r.Revenue
	}

	// ── 2. Hospital bills ──────────────────────────────────────────────────
	hospitalFilter := bson.M{
		"branchId":  branchId,
		"confirm":   true,
		"createdAt": bson.M{"$gte": start, "$lt": end},
	}

	hospPipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: hospitalFilter}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{
					"format":   "%Y-%m-%d",
					"date":     "$createdAt",
					"timezone": "UTC",
				},
			},
			"revenue": bson.M{"$sum": "$totalAmount"},
		}}},
	}

	hospCursor, err := dbConfigs.HospitalBillCollection.Aggregate(ctx, hospPipeline)
	if err != nil {
		return nil, err
	}
	defer hospCursor.Close(ctx)

	var hospRaw []rawRev
	if err = hospCursor.All(ctx, &hospRaw); err != nil {
		return nil, err
	}
	for _, r := range hospRaw {
		revenueMap[r.Date] += r.Revenue
	}

	// Zero-fill: iterate every day in the window
	dates := generateDateSequence(start, end)
	result := make([]dto.RevenueDataPoint, 0, len(dates))
	for _, d := range dates {
		result = append(result, dto.RevenueDataPoint{
			Date:         d,
			TotalRevenue: revenueMap[d],
		})
	}
	return result, nil
}

// ──────────────────────────────────────────────
//  Bonus — Dashboard summary with growth rate
// ──────────────────────────────────────────────

// DB_GetDashboardSummary computes totals for the last `days` period and
// compares them to the previous equal period to derive a revenue growth rate.
func DB_GetDashboardSummary(branchId string, days int) (*dto.DashboardSummary, error) {
	ctx := context.Background()

	// Current period
	curStart, curEnd := buildDayRange(days)
	// Previous period (same length, immediately before current)
	prevEnd := curStart
	prevStart := prevEnd.Add(-time.Duration(days) * 24 * time.Hour)

	// ── Appointment count (current period) ─────────────────────────────────
	apptFilter := bson.M{
		"branchId":        branchId,
		"appointmentDate": bson.M{"$gte": curStart, "$lt": curEnd},
	}
	totalAppts, err := dbConfigs.AppointmentCollection.CountDocuments(ctx, apptFilter)
	if err != nil {
		return nil, err
	}

	// ── Revenue helper ──────────────────────────────────────────────────────
	sumRevenue := func(from, to time.Time) (float64, error) {
		var total float64

		pharmFilter := bson.M{
			"branchId":      branchId,
			"paymentStatus": "PAID",
			"createdAt":     bson.M{"$gte": from, "$lt": to},
		}
		pharmPipeline := mongo.Pipeline{
			bson.D{{Key: "$match", Value: pharmFilter}},
			bson.D{{Key: "$group", Value: bson.M{
				"_id":     nil,
				"revenue": bson.M{"$sum": "$grandTotal"},
			}}},
		}
		pc, err := dbConfigs.BillCollection.Aggregate(ctx, pharmPipeline)
		if err != nil {
			return 0, err
		}
		defer pc.Close(ctx)
		type agg struct{ Revenue float64 `bson:"revenue"` }
		if pc.Next(ctx) {
			var a agg
			if err := pc.Decode(&a); err == nil {
				total += a.Revenue
			}
		}

		hospFilter := bson.M{
			"branchId":  branchId,
			"confirm":   true,
			"createdAt": bson.M{"$gte": from, "$lt": to},
		}
		hospPipeline := mongo.Pipeline{
			bson.D{{Key: "$match", Value: hospFilter}},
			bson.D{{Key: "$group", Value: bson.M{
				"_id":     nil,
				"revenue": bson.M{"$sum": "$totalAmount"},
			}}},
		}
		hc, err := dbConfigs.HospitalBillCollection.Aggregate(ctx, hospPipeline)
		if err != nil {
			return 0, err
		}
		defer hc.Close(ctx)
		if hc.Next(ctx) {
			var a agg
			if err := hc.Decode(&a); err == nil {
				total += a.Revenue
			}
		}

		return total, nil
	}

	curRevenue, err := sumRevenue(curStart, curEnd)
	if err != nil {
		return nil, err
	}
	prevRevenue, err := sumRevenue(prevStart, prevEnd)
	if err != nil {
		return nil, err
	}

	var growthRate float64
	if prevRevenue > 0 {
		growthRate = ((curRevenue - prevRevenue) / prevRevenue) * 100
		// Round to 2 decimal places
		growthRate = float64(int(growthRate*100)) / 100
	}

	return &dto.DashboardSummary{
		TotalAppointments: int(totalAppts),
		TotalRevenue:      curRevenue,
		GrowthRate:        growthRate,
	}, nil
}
