package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ──────────────────────────────────────────────
//  Stock Valuation
// ──────────────────────────────────────────────

type StockValuationItem struct {
	MedicineID      string  `json:"medicineId" bson:"medicineId"`
	MedicineName    string  `json:"medicineName" bson:"medicineName"`
	TotalQty        int     `json:"totalQty" bson:"totalQty"`
	AvgBuyingPrice  float64 `json:"avgBuyingPrice" bson:"avgBuyingPrice"`
	TotalCostValue  float64 `json:"totalCostValue" bson:"totalCostValue"`
	TotalSaleValue  float64 `json:"totalSaleValue" bson:"totalSaleValue"`
}

type StockValuationResponse struct {
	BranchId        string               `json:"branchId"`
	Items           []StockValuationItem `json:"items"`
	GrandCostValue  float64              `json:"grandCostValue"`
	GrandSaleValue  float64              `json:"grandSaleValue"`
}

// ──────────────────────────────────────────────
//  Expiry Alert
// ──────────────────────────────────────────────

type ExpiryAlertItem struct {
	BatchID      string    `json:"batchId" bson:"batchId"`
	MedicineID   string    `json:"medicineId" bson:"medicineId"`
	MedicineName string    `json:"medicineName" bson:"medicineName"`
	BatchNumber  string    `json:"batchNumber" bson:"batchNumber"`
	ExpiryDate   time.Time `json:"expiryDate" bson:"expiryDate"`
	Quantity     int       `json:"quantity" bson:"quantity"`
	DaysToExpiry int       `json:"daysToExpiry" bson:"daysToExpiry"`
	BranchId     string    `json:"branchId" bson:"branchId"`
}

// ──────────────────────────────────────────────
//  Stock Transfer (inter-branch)
// ──────────────────────────────────────────────

type TransferItem struct {
	MedicineID   string `json:"medicineId" bson:"medicineId"`
	MedicineName string `json:"medicineName" bson:"medicineName"`
	BatchId      string `json:"batchId" bson:"batchId"`
	StockId      string `json:"stockId" bson:"stockId"` // FK → BranchStock (source branch)
	BatchNumber  string `json:"batchNumber" bson:"batchNumber"`
	Quantity     int    `json:"quantity" bson:"quantity"`
}

type StockTransferModel struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	TransferId   string             `json:"transferId" bson:"transferId"`
	FromBranchId string             `json:"fromBranchId" bson:"fromBranchId"`
	ToBranchId   string             `json:"toBranchId" bson:"toBranchId"`
	Items        []TransferItem     `json:"items" bson:"items"`
	// PENDING → COMPLETED / CANCELLED
	Status       string    `json:"status" bson:"status"`
	TransferredBy string   `json:"transferredBy" bson:"transferredBy"`
	Notes        string    `json:"notes" bson:"notes"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt" bson:"updatedAt"`
}

type SearchTransferQuery struct {
	FromBranchId string `json:"fromBranchId" query:"fromBranchId"`
	ToBranchId   string `json:"toBranchId" query:"toBranchId"`
	Status       string `json:"status" query:"status"`
	Page         int    `json:"page" query:"page"`
	Limit        int    `json:"limit" query:"limit"`
}
