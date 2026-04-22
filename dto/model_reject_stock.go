package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RejectType categorises the reason for removing stock.
type RejectType = string

const (
	RejectExpired          RejectType = "EXPIRED"
	RejectDamaged          RejectType = "DAMAGED"
	RejectReturnToSupplier RejectType = "RETURN_TO_SUPPLIER"
)

// RejectStockModel represents a request to remove stock from inventory
// due to expiry, damage, or a supplier return.
// Flow: PENDING → APPROVED → COMPLETED
type RejectStockModel struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	RejectId   string             `json:"rejectId" bson:"rejectId"`
	BatchId    string             `json:"batchId" bson:"batchId"` // Global Batch
	StockId    string             `json:"stockId" bson:"stockId"` // Specific Branch Stock
	MedicineId string             `json:"medicineId" bson:"medicineId"`
	BranchId   string             `json:"branchId" bson:"branchId"`
	// Type: EXPIRED | DAMAGED | RETURN_TO_SUPPLIER
	Type     RejectType `json:"type" bson:"type"`
	Quantity int        `json:"quantity" bson:"quantity"`
	Reason   string     `json:"reason" bson:"reason"`
	// Status: PENDING → APPROVED → COMPLETED
	Status     string    `json:"status" bson:"status"`
	ApprovedBy string    `json:"approvedBy,omitempty" bson:"approvedBy,omitempty"`
	ApprovedAt time.Time `json:"approvedAt,omitempty" bson:"approvedAt,omitempty"`
	Notes      string    `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedBy  string    `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	CreatedAt  time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt" bson:"updatedAt"`
}

// SearchRejectQuery filters for listing reject stock entries.
type SearchRejectQuery struct {
	BatchId    string `json:"batchId" query:"batchId"`
	MedicineId string `json:"medicineId" query:"medicineId"`
	BranchId   string `json:"branchId" query:"branchId"`
	Type       string `json:"type" query:"type"`
	Status     string `json:"status" query:"status"`
	From       string `json:"from" query:"from"`
	To         string `json:"to" query:"to"`
	Page       int    `json:"page" query:"page"`
	Limit      int    `json:"limit" query:"limit"`
}
