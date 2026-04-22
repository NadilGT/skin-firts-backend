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

type MedicineBatchModel struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	MedicineBatchId string             `json:"MedicineBatchId" bson:"medicinebatchid"`
	MedicineID      string             `json:"medicineId" bson:"medicineId"`
	Quantity        int                `json:"quantity" bson:"quantity"`
	ExpiryDate      time.Time          `json:"expiryDate" bson:"expiryDate"`
	BuyingPrice     float64            `json:"buyingPrice" bson:"buyingPrice"`
	SellingPrice    float64            `json:"sellingPrice" bson:"sellingPrice"`
	Status           string             `json:"status" bson:"status"` // ACTIVE, OUT_OF_STOCK, EXPIRED
	ReservedQuantity int                `json:"reservedQuantity" bson:"reservedQuantity"`
	// Extended fields
	BatchNumber  string    `json:"batchNumber,omitempty" bson:"batchNumber,omitempty"`
	SupplierId   string    `json:"supplierId,omitempty" bson:"supplierId,omitempty"`
	BranchId     string    `json:"branchId,omitempty" bson:"branchId,omitempty"`
	ReceivedDate time.Time `json:"receivedDate,omitempty" bson:"receivedDate,omitempty"`
	Notes        string    `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
}

type BillItem struct {
	MedicineID string             `json:"medicineId" bson:"medicineId"`
	BatchID    string             `json:"batchId" bson:"batchId"`
	Quantity   int                `json:"quantity" bson:"quantity"`
	Price      float64            `json:"price" bson:"price"`
}

type DeductStockRequest struct {
	MedicineID string `json:"medicineId"`
	Quantity   int    `json:"quantity"`
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
