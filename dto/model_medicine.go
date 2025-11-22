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
	Price                float64            `json:"price" bson:"price"`
	StockQuantity        int                `json:"stockQuantity" bson:"stockQuantity"`
	MinStockLevel        int                `json:"minStockLevel" bson:"minStockLevel"`
	ExpiryDate           string             `json:"expiryDate" bson:"expiryDate"`
	BatchNumber          string             `json:"batchNumber" bson:"batchNumber"`
	Description          string             `json:"description" bson:"description"`
	SideEffects          []string           `json:"sideEffects,omitempty" bson:"sideEffects,omitempty"`
	Contraindications    []string           `json:"contraindications,omitempty" bson:"contraindications,omitempty"`
	PrescriptionRequired bool               `json:"prescriptionRequired" bson:"prescriptionRequired"`
	Status               string             `json:"status" bson:"status"`
	CreatedAt            time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt            time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type SearchMedicineQuery struct {
	SearchTerm string `json:"searchTerm" query:"searchTerm"`
	Category   string `json:"category" query:"category"`
	Form       string `json:"form" query:"form"`
	Status     string `json:"status" query:"status"`
	Page       int    `json:"page" query:"page"`
	Limit      int    `json:"limit" query:"limit"`
}
