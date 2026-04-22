package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MedicineModel struct {
	ID                   primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	MedicineId           string             `json:"MedicineId" bson:"medicineid"`
	Name                 string             `json:"name" bson:"name"`
	GenericName          string             `json:"genericName" bson:"genericName"`
	Manufacturer         string             `json:"manufacturer" bson:"manufacturer"`
	Category             string             `json:"category" bson:"category"`
	Dosage               string             `json:"dosage" bson:"dosage"`
	Form                 string             `json:"form" bson:"form"`
	Strength             string             `json:"strength" bson:"strength"`
	MinStockLevel        int                `json:"minStockLevel" bson:"minStockLevel"`
	// Extended fields
	Barcode              string             `json:"barcode,omitempty" bson:"barcode,omitempty"`
	SupplierId           string             `json:"supplierId,omitempty" bson:"supplierId,omitempty"`
	ReorderLevel         int                `json:"reorderLevel" bson:"reorderLevel"` // alias of minStockLevel for UI clarity
	Description          string             `json:"description" bson:"description"`
	SideEffects          []string           `json:"sideEffects,omitempty" bson:"sideEffects,omitempty"`
	Contraindications    []string           `json:"contraindications,omitempty" bson:"contraindications,omitempty"`
	PrescriptionRequired bool               `json:"prescriptionRequired" bson:"prescriptionRequired"`
	Status               string             `json:"status" bson:"status"`
	CreatedAt            time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt            time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// MedicineBatch is a GLOBAL concept — one shipment from a supplier.
// It has NO quantity, NO branch. It describes what the batch IS.
type MedicineBatch struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	BatchId     string             `json:"batchId" bson:"batchId"`       // BAT-XXXX
	MedicineId  string             `json:"medicineId" bson:"medicineId"` // FK → Medicine
	BatchNumber string             `json:"batchNumber" bson:"batchNumber"` // physical label e.g. "B2024-001"
	ExpiryDate  time.Time          `json:"expiryDate" bson:"expiryDate"`
	BuyingPrice float64            `json:"buyingPrice" bson:"buyingPrice"`
	SellingPrice float64           `json:"sellingPrice" bson:"sellingPrice"`
	SupplierId  string             `json:"supplierId,omitempty" bson:"supplierId,omitempty"`
	// Status: ACTIVE | EXPIRED | BLOCKED (NOT out-of-stock — that's derived from BranchStock)
	Status    string    `json:"status" bson:"status"`
	Notes     string    `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

// BranchStock is a BRANCH-SPECIFIC concept — how much of a batch a branch holds.
// Every branch that holds stock from a given batch gets its own BranchStock document.
type BranchStock struct {
	ID               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	StockId          string             `json:"stockId" bson:"stockId"`     // STK-XXXX
	BatchId          string             `json:"batchId" bson:"batchId"`     // FK → MedicineBatch.BatchId
	MedicineId       string             `json:"medicineId" bson:"medicineId"` // denormalized for query speed
	BranchId         string             `json:"branchId" bson:"branchId"`
	Quantity         int                `json:"quantity" bson:"quantity"`
	ReservedQuantity int                `json:"reservedQuantity" bson:"reservedQuantity"`
	UpdatedAt        time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// BranchStockView is a read-optimised join of MedicineBatch + BranchStock.
// Used for FEFO queries and reports — never written to DB directly.
type BranchStockView struct {
	// From BranchStock
	StockId          string `json:"stockId" bson:"stockId"`
	BatchId          string `json:"batchId" bson:"batchId"`
	MedicineId       string `json:"medicineId" bson:"medicineId"`
	BranchId         string `json:"branchId" bson:"branchId"`
	Quantity         int    `json:"quantity" bson:"quantity"`
	ReservedQuantity int    `json:"reservedQuantity" bson:"reservedQuantity"`
	// From MedicineBatch (joined)
	BatchNumber  string    `json:"batchNumber" bson:"batchNumber"`
	ExpiryDate   time.Time `json:"expiryDate" bson:"expiryDate"`
	SellingPrice float64   `json:"sellingPrice" bson:"sellingPrice"`
	BuyingPrice  float64   `json:"buyingPrice" bson:"buyingPrice"`
	BatchStatus  string    `json:"batchStatus" bson:"batchStatus"`
}

// BillItem represents one line in a confirmed bill.
// BatchID is mandatory (global batch reference).
// StockID is optional but recommended (points to the exact BranchStock doc deducted).
type BillItem struct {
	MedicineID string  `json:"medicineId" bson:"medicineId"`
	BatchID    string  `json:"batchId" bson:"batchId"`
	StockID    string  `json:"stockId,omitempty" bson:"stockId,omitempty"`
	Quantity   int     `json:"quantity" bson:"quantity"`
	Price      float64 `json:"price" bson:"price"`
}

// DeductStockRequest is the payload for a direct stock deduction (e.g. standalone endpoint).
type DeductStockRequest struct {
	MedicineID string `json:"medicineId"`
	Quantity   int    `json:"quantity"`
	BranchId   string `json:"branchId"` // required — stock is branch-specific
}

type SearchMedicineQuery struct {
	SearchTerm string `json:"searchTerm" query:"searchTerm"`
	Category   string `json:"category" query:"category"`
	Form       string `json:"form" query:"form"`
	Status     string `json:"status" query:"status"`
	Page       int    `json:"page" query:"page"`
	Limit      int    `json:"limit" query:"limit"`
}
type CreateBillRequest struct {
	Items         []DeductStockRequest `json:"items"`
	CustomerName  string               `json:"customerName"`
	CustomerPhone string               `json:"customerPhone"`
	Discount      float64              `json:"discount"`
	Tax           float64              `json:"tax"`
	// CASH / CARD / ONLINE
	PaymentMethod string  `json:"paymentMethod"`
	PaidAmount    float64 `json:"paidAmount"`
	BranchId      string  `json:"branchId"`
	Notes         string  `json:"notes"`
	CreatedBy     string  `json:"createdBy"`
}

type BillResponse struct {
	Items              []BillItem `json:"items"`
	TotalMedicinePrice float64    `json:"totalMedicinePrice"`
	AdditionalCharges  float64    `json:"additionalCharges"`
	GrandTotal         float64    `json:"grandTotal"`
}
type BillModel struct {
	ID                 primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	BillId             string             `json:"billId" bson:"billId"`
	PatientID          string             `json:"patientId,omitempty" bson:"patientId,omitempty"`
	Items              []BillItem         `json:"items" bson:"items"`
	TotalMedicinePrice float64            `json:"totalMedicinePrice" bson:"totalMedicinePrice"`
	AdditionalCharges  float64            `json:"additionalCharges" bson:"additionalCharges"`
	GrandTotal         float64            `json:"grandTotal" bson:"grandTotal"`
	Status             string             `json:"status" bson:"status"` // PENDING, CONFIRMED, FAILED
	// Extended POS fields
	CustomerName  string  `json:"customerName,omitempty" bson:"customerName,omitempty"`
	CustomerPhone string  `json:"customerPhone,omitempty" bson:"customerPhone,omitempty"`
	Discount      float64 `json:"discount" bson:"discount"`
	Tax           float64 `json:"tax" bson:"tax"`
	NetTotal      float64 `json:"netTotal" bson:"netTotal"`     // grandTotal - discount + tax
	PaidAmount    float64 `json:"paidAmount" bson:"paidAmount"` // amount actually paid
	DueAmount     float64 `json:"dueAmount" bson:"dueAmount"`   // netTotal - paidAmount
	// PAID / PARTIAL / PENDING
	PaymentStatus string `json:"paymentStatus" bson:"paymentStatus"`
	// CASH / CARD / ONLINE
	PaymentMethod string    `json:"paymentMethod,omitempty" bson:"paymentMethod,omitempty"`
	BranchId      string    `json:"branchId,omitempty" bson:"branchId,omitempty"`
	Notes         string    `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedBy     string    `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	CreatedAt     time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt" bson:"updatedAt"`
}

type UpdateBillPaymentRequest struct {
	PaidAmount    float64 `json:"paidAmount"`
	PaymentMethod string  `json:"paymentMethod"`
	Notes         string  `json:"notes"`
}

type SearchBillQuery struct {
	BranchId      string `json:"branchId" query:"branchId"`
	PaymentStatus string `json:"paymentStatus" query:"paymentStatus"`
	Status        string `json:"status" query:"status"`
	From          string `json:"from" query:"from"`
	To            string `json:"to" query:"to"`
	Page          int    `json:"page" query:"page"`
	Limit         int    `json:"limit" query:"limit"`
}
