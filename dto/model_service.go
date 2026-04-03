package dto

import "go.mongodb.org/mongo-driver/bson/primitive"

// ServiceModel represents a healthcare service offered by the clinic.
type ServiceModel struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ServiceID   string             `json:"serviceId" bson:"serviceId"`
	Name        string             `json:"name"        bson:"name" validate:"required"`
	Description string             `json:"description" bson:"description"`
	UnitPrice   float64            `json:"unitPrice"   bson:"unitPrice"`
}
