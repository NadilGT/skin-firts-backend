package dto

import "go.mongodb.org/mongo-driver/bson/primitive"

type FocusModel struct {
	ID      primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	FocusID string             `json:"focusId" bson:"focusId"`
	Name    string             `json:"name" bson:"name" validate:"required"`
}
