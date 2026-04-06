package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HospitalBillItem represents a single service item within a bill
type HospitalBillItem struct {
	ServiceID   string  `json:"serviceId" bson:"serviceId"`
	ServiceName string  `json:"serviceName" bson:"serviceName"`
	Quantity    int     `json:"quantity" bson:"quantity"`
	UnitPrice   float64 `json:"unitPrice" bson:"unitPrice"`
	Total       float64 `json:"total" bson:"total"`
}

// HospitalBillModel represents a bill generated for multiple hospital services
type HospitalBillModel struct {
	ID             primitive.ObjectID  `json:"id,omitempty" bson:"_id,omitempty"`
	HospitalBillId string              `json:"hospitalBillId" bson:"hospitalBillId"`
	PatientID      string              `json:"patientId" bson:"patientId"`
	PatientName    string              `json:"patientName" bson:"patientName"`
	DoctorID       string              `json:"doctorId" bson:"doctorId"`
	DoctorName     string              `json:"doctorName" bson:"doctorName"`
	Items          []HospitalBillItem  `json:"items" bson:"items"`
	TotalAmount    float64             `json:"totalAmount" bson:"totalAmount"`
	Confirm        bool                `json:"confirm" bson:"confirm"`
	CreatedAt      time.Time           `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time           `json:"updatedAt" bson:"updatedAt"`
}

// BillServiceItem is the input payload for each service
type BillServiceItem struct {
	ServiceID string `json:"serviceId" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required"`
}

// CreateHospitalBillRequest represents the incoming request payload for bill creation
type CreateHospitalBillRequest struct {
	Items       []BillServiceItem `json:"items" validate:"required,min=1"`
	PatientID   string            `json:"patientId"`
	PatientName string            `json:"patientName"`
	DoctorID    string            `json:"doctorId"`
	DoctorName  string            `json:"doctorName"`
}
