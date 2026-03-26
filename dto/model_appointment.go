package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AppointmentModel struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	AppointmentID   string             `json:"appointmentId" bson:"appointmentId"`
	AppointmentNumber int              `json:"appointmentNumber" bson:"appointmentNumber"`
	PatientID       string             `json:"patientId" bson:"patientId"`
	PatientName     string             `json:"patientName" bson:"patientName" validate:"required"`
	PatientEmail    string             `json:"patientEmail" bson:"patientEmail" validate:"required,email"`
	PatientPhone    string             `json:"patientPhone" bson:"patientPhone"`
	DoctorID        string             `json:"doctorId" bson:"doctorId" validate:"required"`
	DoctorName      string             `json:"doctorName" bson:"doctorName" validate:"required"`
	DoctorSpecialty string             `json:"doctorSpecialty" bson:"doctorSpecialty"`
	AppointmentDate time.Time          `json:"appointmentDate" bson:"appointmentDate" validate:"required"`
	Status          string             `json:"status" bson:"status"` // confirmed, pending, completed, cancelled
	Notes           string             `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt       time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type CreateAppointmentRequest struct {
	AppointmentID     string `json:"appointmentId" bson:"appointmentId"`
	AppointmentNumber int    `json:"appointmentNumber" bson:"appointmentNumber"`
	PatientID         string `json:"patientId" bson:"patientId"`
	PatientName       string `json:"patientName" bson:"patientName" validate:"required"`
	PatientEmail      string `json:"patientEmail" bson:"patientEmail" validate:"required,email"`
	PatientPhone      string `json:"patientPhone" bson:"patientPhone"`
	DoctorID          string `json:"doctorId" bson:"doctorId" validate:"required"`
	DoctorName        string `json:"doctorName" bson:"doctorName" validate:"required"`
	DoctorSpecialty   string `json:"doctorSpecialty" bson:"doctorSpecialty"`
	AppointmentDate   string `json:"appointmentDate" bson:"appointmentDate" validate:"required"` // Format: 2025-11-09
	Notes             string `json:"notes,omitempty" bson:"notes,omitempty"`
}

type UpdateAppointmentStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=confirmed pending completed cancelled running"`
}
