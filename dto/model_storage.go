package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Rack struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	RackId      string             `json:"rackId" bson:"rackId"`
	Name        string             `json:"name" bson:"name"` // e.g. A, B, C
	Description string             `json:"description" bson:"description"`
	IsActive    bool               `json:"isActive" bson:"isActive"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type Shelf struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ShelfId     string             `json:"shelfId" bson:"shelfId"`
	RackId      string             `json:"rackId" bson:"rackId"`
	Name        string             `json:"name" bson:"name"` // e.g. 1, 2, 3
	Description string             `json:"description" bson:"description"`
	IsActive    bool               `json:"isActive" bson:"isActive"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type Location struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	LocationId  string             `json:"locationId" bson:"locationId"`
	RackId      string             `json:"rackId" bson:"rackId"`
	ShelfId     string             `json:"shelfId" bson:"shelfId"`
	Position    int                `json:"position" bson:"position"`
	Code        string             `json:"code" bson:"code"` // AUTO GENERATED, IMMUTABLE
	Description string             `json:"description" bson:"description"`
	IsOccupied  bool               `json:"isOccupied" bson:"isOccupied"`
	IsActive    bool               `json:"isActive" bson:"isActive"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
}
