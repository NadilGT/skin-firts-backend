package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ──────────────────────────────────────────────
//  Storage Dashboard / Warehouse Map DTOs
// ──────────────────────────────────────────────

// BranchStockSlim is a lightweight per-branch stock summary attached to a batch
// when rendering the warehouse map (avoids returning full BranchStock docs).
type BranchStockSlim struct {
	BranchId  string `json:"branchId"`
	Quantity  int    `json:"quantity"`
	Available int    `json:"available"` // quantity - reservedQuantity
}

// BatchSummary is the lightweight batch info shown inside a location slot
// on the rack map UI.
type BatchSummary struct {
	BatchId      string           `json:"batchId"`
	MedicineId   string           `json:"medicineId"`
	MedicineName string           `json:"medicineName"`
	BatchNumber  string           `json:"batchNumber"`
	ExpiryDate   time.Time        `json:"expiryDate"`
	Status       string           `json:"status"`
	BranchStocks []BranchStockSlim `json:"branchStocks"`
}

// LocationWithBatches represents a single location slot enriched with its batches.
type LocationWithBatches struct {
	LocationId  string        `json:"locationId"`
	Code        string        `json:"code"`
	Position    int           `json:"position"`
	Description string        `json:"description"`
	IsOccupied  bool          `json:"isOccupied"`
	IsActive    bool          `json:"isActive"`
	Batches     []BatchSummary `json:"batches"`
}

// ShelfWithLocations represents a shelf enriched with its location slots.
type ShelfWithLocations struct {
	ShelfId     string               `json:"shelfId"`
	RackId      string               `json:"rackId"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	IsActive    bool                 `json:"isActive"`
	Locations   []LocationWithBatches `json:"locations"`
}

// RackWithShelves is the top-level node in the warehouse map tree.
type RackWithShelves struct {
	RackId      string               `json:"rackId"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	IsActive    bool                 `json:"isActive"`
	Shelves     []ShelfWithLocations `json:"shelves"`
}

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
