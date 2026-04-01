package dto

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReportModel struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ReportID      string             `json:"reportId" bson:"reportId"`
	PatientID     string             `json:"patientId" bson:"patientId"`
	AppointmentID string             `json:"appointmentId" bson:"appointmentId"`
	Title         string             `json:"title" bson:"title"`
	Description   string             `json:"description" bson:"description"`
	FileURL       string             `json:"fileUrl" bson:"fileUrl"`
	FileType      string             `json:"fileType" bson:"fileType"`
	UploadedBy    string             `json:"uploadedBy" bson:"uploadedBy"`
	Status        string             `json:"status" bson:"status"`
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
}