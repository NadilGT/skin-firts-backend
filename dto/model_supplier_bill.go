package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SupplierBillLineItem is a single line in a supplier invoice.
type SupplierBillLineItem struct {
	MedicineId   string  `json:"medicineId" bson:"medicineId"`
	MedicineName string  `json:"medicineName" bson:"medicineName"`
	BatchNumber  string  `json:"batchNumber,omitempty" bson:"batchNumber,omitempty"`
	Quantity     int     `json:"quantity" bson:"quantity"`
	UnitCost     float64 `json:"unitCost" bson:"unitCost"`
	TotalCost    float64 `json:"totalCost" bson:"totalCost"`
}

// SupplierBillModel is the financial record from the supplier (invoice).
// Linked to PO and/or GRN.
// Standard flow: PO → GRN → SupplierBill
type SupplierBillModel struct {
	ID              primitive.ObjectID     `json:"id,omitempty" bson:"_id,omitempty"`
	BillId          string                 `json:"billId" bson:"billId"`
	SupplierId      string                 `json:"supplierId" bson:"supplierId"`
	SupplierName    string                 `json:"supplierName" bson:"supplierName"`
	// Optional links to PO and GRN
	PurchaseOrderId string                 `json:"purchaseOrderId,omitempty" bson:"purchaseOrderId,omitempty"`
	GrnId           string                 `json:"grnId,omitempty" bson:"grnId,omitempty"`
	BranchId        string                 `json:"branchId" bson:"branchId"`
	Items           []SupplierBillLineItem `json:"items" bson:"items"`
	TotalAmount     float64                `json:"totalAmount" bson:"totalAmount"`
	PaidAmount      float64                `json:"paidAmount" bson:"paidAmount"`
	DueAmount       float64                `json:"dueAmount" bson:"dueAmount"`
	// PaymentStatus: UNPAID | PARTIAL | PAID
	PaymentStatus string    `json:"paymentStatus" bson:"paymentStatus"`
	// PaymentMethod: CASH | BANK_TRANSFER | CHEQUE
	PaymentMethod string    `json:"paymentMethod,omitempty" bson:"paymentMethod,omitempty"`
	DueDate       time.Time `json:"dueDate,omitempty" bson:"dueDate,omitempty"`
	Notes         string    `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedBy     string    `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	CreatedAt     time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt" bson:"updatedAt"`
}

// UpdateSupplierBillPaymentRequest updates the bill's payment fields.
type UpdateSupplierBillPaymentRequest struct {
	PaidAmount    float64 `json:"paidAmount"`
	PaymentMethod string  `json:"paymentMethod"`
	Notes         string  `json:"notes"`
}

// SearchSupplierBillQuery filters for listing supplier bills.
type SearchSupplierBillQuery struct {
	SupplierId      string `json:"supplierId" query:"supplierId"`
	BranchId        string `json:"branchId" query:"branchId"`
	PurchaseOrderId string `json:"purchaseOrderId" query:"purchaseOrderId"`
	GrnId           string `json:"grnId" query:"grnId"`
	PaymentStatus   string `json:"paymentStatus" query:"paymentStatus"`
	From            string `json:"from" query:"from"`
	To              string `json:"to" query:"to"`
	Page            int    `json:"page" query:"page"`
	Limit           int    `json:"limit" query:"limit"`
}
