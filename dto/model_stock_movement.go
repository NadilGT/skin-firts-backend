package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MovementType represents the kind of inventory change.
type MovementType = string

const (
	MovementPurchase     MovementType = "PURCHASE"
	MovementSale         MovementType = "SALE"
	MovementTransferIn   MovementType = "TRANSFER_IN"
	MovementTransferOut  MovementType = "TRANSFER_OUT"
	MovementReject       MovementType = "REJECT"
	MovementAdjustment   MovementType = "ADJUSTMENT"
)

// StockMovementModel is the immutable audit ledger entry.
// Every change to batch quantity (GRN, sale, transfer, reject, adjustment)
// must create a corresponding StockMovement.
type StockMovementModel struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	MovementId    string             `json:"movementId" bson:"movementId"`
	BatchId       string             `json:"batchId" bson:"batchId"`
	MedicineId    string             `json:"medicineId" bson:"medicineId"`
	BranchId      string             `json:"branchId" bson:"branchId"`
	// Type: PURCHASE | SALE | TRANSFER_IN | TRANSFER_OUT | REJECT | ADJUSTMENT
	Type          MovementType `json:"type" bson:"type"`
	// Quantity is always positive. Direction is inferred from Type.
	// PURCHASE, TRANSFER_IN, ADJUSTMENT(+) → stock increases
	// SALE, TRANSFER_OUT, REJECT, ADJUSTMENT(-) → stock decreases
	Quantity      int    `json:"quantity" bson:"quantity"`
	// ReferenceId links to PO / GRN / Bill / Transfer / RejectStock document id
	ReferenceId   string `json:"referenceId,omitempty" bson:"referenceId,omitempty"`
	// ReferenceType: PO | GRN | BILL | TRANSFER | REJECT | MANUAL
	ReferenceType string `json:"referenceType,omitempty" bson:"referenceType,omitempty"`
	Notes         string `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedBy     string `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	CreatedAt     time.Time `json:"createdAt" bson:"createdAt"`
}

// SearchMovementQuery are the filters for listing stock movements.
type SearchMovementQuery struct {
	BatchId       string `json:"batchId" query:"batchId"`
	MedicineId    string `json:"medicineId" query:"medicineId"`
	BranchId      string `json:"branchId" query:"branchId"`
	Type          string `json:"type" query:"type"`
	ReferenceId   string `json:"referenceId" query:"referenceId"`
	ReferenceType string `json:"referenceType" query:"referenceType"`
	From          string `json:"from" query:"from"`
	To            string `json:"to" query:"to"`
	Page          int    `json:"page" query:"page"`
	Limit         int    `json:"limit" query:"limit"`
}
