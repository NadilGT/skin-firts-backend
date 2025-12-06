package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AppointmentModel struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	PatientID      string             `json:"patientId" bson:"patientId" validate:"required"`
	PatientName    string             `json:"patientName" bson:"patientName" validate:"required"`
	PatientEmail   string             `json:"patientEmail" bson:"patientEmail" validate:"required,email"`
	PatientPhone   string             `json:"patientPhone" bson:"patientPhone"`
	DoctorID       string             `json:"doctorId" bson:"doctorId" validate:"required"`
	DoctorName     string             `json:"doctorName" bson:"doctorName" validate:"required"`
	DoctorSpecialty string            `json:"doctorSpecialty" bson:"doctorSpecialty"`
	AppointmentDate time.Time         `json:"appointmentDate" bson:"appointmentDate" validate:"required"`
	TimeSlot       string             `json:"timeSlot" bson:"timeSlot" validate:"required"`
	Status         string             `json:"status" bson:"status"` // confirmed, pending, completed, cancelled
	Notes          string             `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt      time.Time         `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt" bson:"updatedAt"`
}

type CreateAppointmentRequest struct {
	PatientID       string `json:"patientId" validate:"required"`
	PatientName     string `json:"patientName" validate:"required"`
	PatientEmail    string `json:"patientEmail" validate:"required,email"`
	PatientPhone    string `json:"patientPhone"`
	DoctorID        string `json:"doctorId" validate:"required"`
	DoctorName      string `json:"doctorName" validate:"required"`
	DoctorSpecialty string `json:"doctorSpecialty"`
	AppointmentDate string `json:"appointmentDate" validate:"required"` // Format: 2025-11-09
	TimeSlot        string `json:"timeSlot" validate:"required"`       // Format: 10:00 AM
	Notes           string `json:"notes,omitempty"`
}

type UpdateAppointmentStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=confirmed pending completed cancelled"`
}