package dto

// ──────────────────────────────────────────────
//  Dashboard Analytics — Time-Series DTOs
// ──────────────────────────────────────────────

// AppointmentDataPoint is a single day's appointment count for chart rendering.
// { "date": "2026-05-01", "count": 12 }
type AppointmentDataPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// RevenueDataPoint is a single day's total revenue for chart rendering.
// { "date": "2026-05-01", "totalRevenue": 1200 }
type RevenueDataPoint struct {
	Date         string  `json:"date"`
	TotalRevenue float64 `json:"totalRevenue"`
}

// DashboardSummary is the aggregated summary for the dashboard header cards.
type DashboardSummary struct {
	TotalAppointments int     `json:"totalAppointments"`
	TotalRevenue      float64 `json:"totalRevenue"`
	// GrowthRate is the percentage change in revenue vs the previous equal period.
	// Positive = growth, negative = decline. 0 when no prior period data exists.
	GrowthRate float64 `json:"growthRate"`
}
