package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ──────────────────────────────────────────────
//  Supplier
// ──────────────────────────────────────────────

type SupplierModel struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	SupplierId    string             `json:"supplierId" bson:"supplierId"`
	Name          string             `json:"name" bson:"name"`
	ContactPerson string             `json:"contactPerson" bson:"contactPerson"`
	Phone         string             `json:"phone" bson:"phone"`
	Email         string             `json:"email" bson:"email"`
	Address       string             `json:"address" bson:"address"`
	TaxNo         string             `json:"taxNo" bson:"taxNo"`
	PaymentTerms  string             `json:"paymentTerms" bson:"paymentTerms"` // e.g. "Net 30"
	Status        string             `json:"status" bson:"status"`             // ACTIVE / INACTIVE
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// ──────────────────────────────────────────────
//  Purchase Order
// ──────────────────────────────────────────────

type POLineItem struct {
	MedicineID   string  `json:"medicineId" bson:"medicineId"`
	MedicineName string  `json:"medicineName" bson:"medicineName"`
	Quantity     int     `json:"quantity" bson:"quantity"`
	UnitCost     float64 `json:"unitCost" bson:"unitCost"`
	TotalCost    float64 `json:"totalCost" bson:"totalCost"`
}

type PurchaseOrderModel struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	PoId         string             `json:"poId" bson:"poId"`
	SupplierId   string             `json:"supplierId" bson:"supplierId"`
	SupplierName string             `json:"supplierName" bson:"supplierName"`
	BranchId     string             `json:"branchId" bson:"branchId"`
	Items        []POLineItem       `json:"items" bson:"items"`
	TotalAmount  float64            `json:"totalAmount" bson:"totalAmount"`
	// DRAFT → ORDERED → PARTIAL → RECEIVED → CANCELLED
	Status       string    `json:"status" bson:"status"`
	ExpectedDate time.Time `json:"expectedDate" bson:"expectedDate"`
	Notes        string    `json:"notes" bson:"notes"`
	CreatedBy    string    `json:"createdBy" bson:"createdBy"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt" bson:"updatedAt"`
}

type UpdatePOStatusRequest struct {
	Status string `json:"status"` // APPROVED & ORDERED & PARTIALLY_RECEIVED
	Notes  string `json:"notes"`
}

// ──────────────────────────────────────────────
//  GRN — Goods Received Note
// ──────────────────────────────────────────────

type GRNLineItem struct {
	MedicineID   string    `json:"medicineId" bson:"medicineId"`
	MedicineName string    `json:"medicineName" bson:"medicineName"`
	BatchNumber  string    `json:"batchNumber" bson:"batchNumber"`
	Quantity     int       `json:"quantity" bson:"quantity"`
	ExpiryDate   time.Time `json:"expiryDate" bson:"expiryDate"`
	BuyingPrice  float64   `json:"buyingPrice" bson:"buyingPrice"`
	SellingPrice float64   `json:"sellingPrice" bson:"sellingPrice"`
}

type GRNModel struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	GrnId        string             `json:"grnId" bson:"grnId"`
	PoId         string             `json:"poId,omitempty" bson:"poId,omitempty"` // optional link to PO
	SupplierId   string             `json:"supplierId" bson:"supplierId"`
	SupplierName string             `json:"supplierName" bson:"supplierName"`
	BranchId     string             `json:"branchId" bson:"branchId"`
	Items        []GRNLineItem      `json:"items" bson:"items"`
	ReceivedDate time.Time          `json:"receivedDate" bson:"receivedDate"`
	ReceivedBy   string             `json:"receivedBy" bson:"receivedBy"`
	Notes        string             `json:"notes" bson:"notes"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
}

type SearchSupplierQuery struct {
	SearchTerm string `json:"searchTerm" query:"searchTerm"`
	Status     string `json:"status" query:"status"`
	Page       int    `json:"page" query:"page"`
	Limit      int    `json:"limit" query:"limit"`
}

type SearchPOQuery struct {
	SupplierId string `json:"supplierId" query:"supplierId"`
	BranchId   string `json:"branchId" query:"branchId"`
	Status     string `json:"status" query:"status"`
	Page       int    `json:"page" query:"page"`
	Limit      int    `json:"limit" query:"limit"`
}

type SearchGRNQuery struct {
	SupplierId string `json:"supplierId" query:"supplierId"`
	BranchId   string `json:"branchId" query:"branchId"`
	PoId       string `json:"poId" query:"poId"`
	Page       int    `json:"page" query:"page"`
	Limit      int    `json:"limit" query:"limit"`
}
