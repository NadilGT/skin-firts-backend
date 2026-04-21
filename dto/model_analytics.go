package dto

// ──────────────────────────────────────────────
//  Top Selling Medicines
// ──────────────────────────────────────────────

type TopSellingItem struct {
	MedicineID   string  `json:"medicineId" bson:"medicineId"`
	MedicineName string  `json:"medicineName" bson:"medicineName"`
	TotalQtySold int     `json:"totalQtySold" bson:"totalQtySold"`
	TotalRevenue float64 `json:"totalRevenue" bson:"totalRevenue"`
}

// ──────────────────────────────────────────────
//  Sales Report (daily / monthly)
// ──────────────────────────────────────────────

type SalesReportItem struct {
	Period       string  `json:"period" bson:"period"` // "2025-04-20" or "2025-04"
	TotalBills   int     `json:"totalBills" bson:"totalBills"`
	TotalRevenue float64 `json:"totalRevenue" bson:"totalRevenue"`
	TotalDiscount float64 `json:"totalDiscount" bson:"totalDiscount"`
}

// ──────────────────────────────────────────────
//  Profit Margin Report
// ──────────────────────────────────────────────

type ProfitMarginItem struct {
	MedicineID    string  `json:"medicineId" bson:"medicineId"`
	MedicineName  string  `json:"medicineName" bson:"medicineName"`
	TotalQtySold  int     `json:"totalQtySold" bson:"totalQtySold"`
	TotalRevenue  float64 `json:"totalRevenue" bson:"totalRevenue"`
	TotalCost     float64 `json:"totalCost" bson:"totalCost"`
	GrossProfit   float64 `json:"grossProfit" bson:"grossProfit"`
	ProfitMarginPct float64 `json:"profitMarginPct" bson:"profitMarginPct"`
}

// ──────────────────────────────────────────────
//  Stock Report
// ──────────────────────────────────────────────

type StockReportItem struct {
	MedicineID   string  `json:"medicineId" bson:"medicineId"`
	MedicineName string  `json:"medicineName" bson:"medicineName"`
	Category     string  `json:"category" bson:"category"`
	TotalQty     int     `json:"totalQty" bson:"totalQty"`
	ReorderLevel int     `json:"reorderLevel" bson:"reorderLevel"`
	IsLowStock   bool    `json:"isLowStock" bson:"isLowStock"`
	TotalBatches int     `json:"totalBatches" bson:"totalBatches"`
}

// ──────────────────────────────────────────────
//  Query params
// ──────────────────────────────────────────────

type AnalyticsQuery struct {
	BranchId string `json:"branchId" query:"branchId"`
	From     string `json:"from" query:"from"` // YYYY-MM-DD
	To       string `json:"to" query:"to"`
	Period   string `json:"period" query:"period"` // "daily" | "monthly"
	Limit    int    `json:"limit" query:"limit"`
	Days     int    `json:"days" query:"days"` // for expiry alerts
}
