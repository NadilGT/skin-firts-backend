package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BranchModel struct {
	ID           *primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	BranchId     string              `json:"branchId,omitempty" bson:"branchId,omitempty"`
	Name         string              `json:"name" bson:"name"`
	Address      string              `json:"address" bson:"address"`
	Phone        string              `json:"phone" bson:"phone"`
	Email        string              `json:"email" bson:"email"`
	IsMainBranch bool                `json:"isMainBranch" bson:"isMainBranch"`
	Status       string              `json:"status" bson:"status"` // ACTIVE / INACTIVE
	CreatedAt    *time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt    *time.Time          `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}
