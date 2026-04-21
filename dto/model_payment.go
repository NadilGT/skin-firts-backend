package dto

// ──────────────────────────────────────────────
//  Daily Sales Summary
// ──────────────────────────────────────────────

type DailySalesSummary struct {
	Date          string  `json:"date"`
	BranchId      string  `json:"branchId"`
	TotalBills    int     `json:"totalBills"`
	TotalRevenue  float64 `json:"totalRevenue"`
	CashRevenue   float64 `json:"cashRevenue"`
	CardRevenue   float64 `json:"cardRevenue"`
	OnlineRevenue float64 `json:"onlineRevenue"`
	TotalDiscount float64 `json:"totalDiscount"`
	TotalTax      float64 `json:"totalTax"`
	TotalDue      float64 `json:"totalDue"`
}

// ──────────────────────────────────────────────
//  Revenue Summary (date range)
// ──────────────────────────────────────────────

type RevenueSummaryItem struct {
	Period        string  `json:"period"`        // "2025-04-20" or "2025-04"
	TotalBills    int     `json:"totalBills"`
	TotalRevenue  float64 `json:"totalRevenue"`
	TotalDiscount float64 `json:"totalDiscount"`
	TotalTax      float64 `json:"totalTax"`
}

type RevenueSummaryResponse struct {
	BranchId  string               `json:"branchId"`
	From      string               `json:"from"`
	To        string               `json:"to"`
	Items     []RevenueSummaryItem `json:"items"`
	GrandTotal float64             `json:"grandTotal"`
}

// ──────────────────────────────────────────────
//  Query params
// ──────────────────────────────────────────────

type PaymentSummaryQuery struct {
	BranchId      string `json:"branchId" query:"branchId"`
	Date          string `json:"date" query:"date"` // YYYY-MM-DD
	From          string `json:"from" query:"from"`
	To            string `json:"to" query:"to"`
	PaymentStatus string `json:"paymentStatus" query:"paymentStatus"`
	Page          int    `json:"page" query:"page"`
	Limit         int    `json:"limit" query:"limit"`
}
