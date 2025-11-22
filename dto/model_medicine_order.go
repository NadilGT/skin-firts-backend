package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderLineItem struct {
	MedicineID   string  `json:"medicineId" bson:"medicineid"` // The custom MedicineId, e.g., "MED-001"
	Name         string  `json:"name" bson:"name"`             // Snapshot of the medicine name
	Quantity     int     `json:"quantity" bson:"quantity"`
	UnitQuantity int     `json:"unitQuantity" bson:"unitQuantity"` // Stock units (e.g., 1 box has 10 strips)
	UnitCost     float64 `json:"unitCost" bson:"unitCost"`     // Price per unit at the time of order
	TotalPrice   float64 `json:"totalPrice" bson:"totalPrice"` // Quantity * UnitCost
}

type MedicineOrderModel struct {
	ID                primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ReceiptID         string             `json:"receiptId" bson:"receiptid"` // Unique ID like "ORD-2025-0001"
	PatientID         string             `json:"patientId,omitempty" bson:"patientId,omitempty"` // ID of registered patient
	PatientName       string             `json:"patientName" bson:"patientName"`
	ContactNumber     string             `json:"contactNumber" bson:"contactNumber"`
	ShippingAddress   string             `json:"shippingAddress" bson:"shippingAddress"`
	
	Items             []OrderLineItem    `json:"items" bson:"items"`
	TotalAmount       float64            `json:"totalAmount" bson:"totalAmount"`
	OrderStatus       string             `json:"orderStatus" bson:"orderStatus"` // e.g., "Pending", "Processing", "Dispensed", "Delivered", "Canceled"
	PrescriptionRef   string             `json:"prescriptionRef,omitempty" bson:"prescriptionRef,omitempty"` // Link to a digital prescription
	DispensedByUserID string             `json:"dispensedByUserId,omitempty" bson:"dispensedByUserId,omitempty"`
	
	CreatedAt         time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt         time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type SearchOrderQuery struct {
	ReceiptID string `json:"receiptId" query:"receiptId"`
	PatientName string `json:"patientName" query:"patientName"`
	ContactNumber string `json:"contactNumber" query:"contactNumber"`
	Status string `json:"status" query:"status"`
	Page int `json:"page" query:"page"`
	Limit int `json:"limit" query:"limit"`
}

type UpdateOrderStatusRequest struct {
	OrderStatus       string `json:"orderStatus" bson:"orderStatus"`
	DispensedByUserID string `json:"dispensedByUserId,omitempty" bson:"dispensedByUserId,omitempty"`
}