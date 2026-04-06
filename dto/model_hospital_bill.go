package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HospitalBillModel represents a bill generated for a specific hospital service
type HospitalBillModel struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	HospitalBillId string             `json:"hospitalBillId" bson:"hospitalBillId"`
	PatientID      string             `json:"patientId" bson:"patientId"`
	PatientName    string             `json:"patientName" bson:"patientName"`
	DoctorID       string             `json:"doctorId" bson:"doctorId"`
	DoctorName     string             `json:"doctorName" bson:"doctorName"`
	ServiceID      string             `json:"serviceId" bson:"serviceId"`
	ServiceName    string             `json:"serviceName" bson:"serviceName"`
	Quantity       int                `json:"quantity" bson:"quantity"`
	UnitPrice      float64            `json:"unitPrice" bson:"unitPrice"`
	TotalAmount    float64            `json:"totalAmount" bson:"totalAmount"`
	Confirm        bool               `json:"confirm" bson:"confirm"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// CreateHospitalBillRequest represents the incoming request payload for bill creation
type CreateHospitalBillRequest struct {
	ServiceID   string `json:"serviceId" validate:"required"`
	Quantity    int    `json:"quantity" validate:"required"`
	PatientID   string `json:"patientId"`
	PatientName string `json:"patientName"`
	DoctorID    string `json:"doctorId"`
	DoctorName  string `json:"doctorName"`
}
