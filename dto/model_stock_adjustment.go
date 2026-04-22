package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockAdjustmentModel struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	AdjustmentId string             `json:"adjustmentId" bson:"adjustmentId"`
	BatchId      string             `json:"batchId" bson:"batchId"`
	MedicineId   string             `json:"medicineId" bson:"medicineId"`
	BranchId     string             `json:"branchId" bson:"branchId"`
	Type         string             `json:"type" bson:"type"` // ADJUSTMENT_IN or ADJUSTMENT_OUT
	Quantity     int                `json:"quantity" bson:"quantity"`
	Reason       string             `json:"reason" bson:"reason"`
	Notes        string             `json:"notes,omitempty" bson:"notes,omitempty"`
	Status       string             `json:"status" bson:"status"` // PENDING, APPROVED, COMPLETED
	CreatedBy    string             `json:"createdBy" bson:"createdBy"`
	ApprovedBy   string             `json:"approvedBy,omitempty" bson:"approvedBy,omitempty"`
	ApprovedAt   time.Time          `json:"approvedAt,omitempty" bson:"approvedAt,omitempty"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type SearchAdjustmentQuery struct {
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
