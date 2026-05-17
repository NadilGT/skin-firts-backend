package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// ==========================================
// RACK APIs
// ==========================================

func CreateRack(c *fiber.Ctx) error {
	var rack dto.Rack
	if err := c.BodyParser(&rack); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body: " + err.Error()})
	}
	if rack.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
	}
	branchId, err := ResolveBranchId(c, rack.BranchId)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized or missing branchId"})
	}
	rack.BranchId = branchId
	rack.IsActive = true

	if err := dao.DB_CreateRack(rack); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create rack: " + err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Rack created successfully", "data": rack})
}

func GetRacks(c *fiber.Ctx) error {
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized or missing branchId"})
	}
	search := c.Query("search")
	activeOnly := c.Query("activeOnly") == "true"
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	racks, total, err := dao.DB_GetRacks(branchId, search, activeOnly, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch racks"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": racks,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func GetRackByID(c *fiber.Ctx) error {
	id := c.Params("id")
	rack, err := dao.DB_GetRackByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Rack not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": rack})
}

func UpdateRack(c *fiber.Ctx) error {
	id := c.Params("id")
	var rack dto.Rack
	if err := c.BodyParser(&rack); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	rack.RackId = id

	if err := dao.DB_UpdateRack(rack); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update rack: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Rack updated successfully"})
}

func DeactivateRack(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := dao.DB_DeactivateRack(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Rack deactivated"})
}

func ActivateRack(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := dao.DB_ActivateRack(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Rack activated"})
}

// ==========================================
// SHELF APIs
// ==========================================

func CreateShelf(c *fiber.Ctx) error {
	var shelf dto.Shelf
	if err := c.BodyParser(&shelf); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if shelf.Name == "" || shelf.RackId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name and RackId are required"})
	}
	branchId, err := ResolveBranchId(c, shelf.BranchId)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized or missing branchId"})
	}
	shelf.BranchId = branchId
	shelf.IsActive = true

	if err := dao.DB_CreateShelf(shelf); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create shelf: " + err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Shelf created successfully", "data": shelf})
}

func GetShelves(c *fiber.Ctx) error {
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized or missing branchId"})
	}
	rackId := c.Query("rackId")
	activeOnly := c.Query("activeOnly") == "true"
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	shelves, total, err := dao.DB_GetShelves(branchId, rackId, activeOnly, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch shelves"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": shelves,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func GetShelfByID(c *fiber.Ctx) error {
	id := c.Params("id")
	shelf, err := dao.DB_GetShelfByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shelf not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": shelf})
}

func UpdateShelf(c *fiber.Ctx) error {
	id := c.Params("id")
	var shelf dto.Shelf
	if err := c.BodyParser(&shelf); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	shelf.ShelfId = id

	if err := dao.DB_UpdateShelf(shelf); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update shelf: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Shelf updated successfully"})
}

func DeactivateShelf(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := dao.DB_DeactivateShelf(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Shelf deactivated"})
}

func ActivateShelf(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := dao.DB_ActivateShelf(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Shelf activated"})
}

func GetShelvesByRackID(c *fiber.Ctx) error {
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized or missing branchId"})
	}
	rackId := c.Params("rackId")
	activeOnly := c.Query("activeOnly") == "true"
	
	// Default page and limit for fetching all inside a rack
	page := 1
	limit := 1000 

	shelves, _, err := dao.DB_GetShelves(branchId, rackId, activeOnly, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch shelves"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": shelves})
}

// ==========================================
// LOCATION APIs
// ==========================================

func CreateLocation(c *fiber.Ctx) error {
	var location dto.Location
	if err := c.BodyParser(&location); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if location.RackId == "" || location.ShelfId == "" || location.Position <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "RackId, ShelfId, and valid Position (>0) are required"})
	}
	branchId, err := ResolveBranchId(c, location.BranchId)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized or missing branchId"})
	}
	location.BranchId = branchId
	location.IsActive = true
	location.IsOccupied = false

	if err := dao.DB_CreateLocation(location); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to create location: " + err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Location created successfully", "data": location})
}

func GetLocations(c *fiber.Ctx) error {
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized or missing branchId"})
	}
	rackId := c.Query("rackId")
	shelfId := c.Query("shelfId")
	searchCode := c.Query("code")
	activeOnly := c.Query("activeOnly") == "true"
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	locations, total, err := dao.DB_GetLocations(branchId, rackId, shelfId, searchCode, activeOnly, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch locations"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": locations,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func GetLocationByID(c *fiber.Ctx) error {
	id := c.Params("id")
	location, err := dao.DB_GetLocationByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Location not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": location})
}

func UpdateLocation(c *fiber.Ctx) error {
	id := c.Params("id")
	var location dto.Location
	if err := c.BodyParser(&location); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	location.LocationId = id

	if err := dao.DB_UpdateLocation(location); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update location: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Location updated successfully"})
}

func DeactivateLocation(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := dao.DB_DeactivateLocation(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Location deactivated"})
}

func ActivateLocation(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := dao.DB_ActivateLocation(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Location activated"})
}

func GetLocationsByShelfID(c *fiber.Ctx) error {
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized or missing branchId"})
	}
	shelfId := c.Params("shelfId")
	activeOnly := c.Query("activeOnly") == "true"
	
	// Fetch all locations for shelf
	page := 1
	limit := 1000 

	locations, _, err := dao.DB_GetLocations(branchId, "", shelfId, "", activeOnly, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch locations"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": locations})
}

// ==========================================
// WAREHOUSE MAP / DASHBOARD APIs
// ==========================================

// GetWarehouseMap returns the complete storage hierarchy tree:
//   Rack → Shelf → Location → [Batches with per-branch stock quantities]
//
// This is the primary endpoint for the visual rack/shelf/location dashboard.
// The frontend can render each rack as a grid of shelf rows × location columns,
// coloring each cell by occupancy / expiry status.
//
// GET /api/warehouse/map
func GetWarehouseMap(c *fiber.Ctx) error {
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized or missing branchId"})
	}
	tree, err := dao.DB_GetWarehouseMap(branchId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to build warehouse map: " + err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":  tree,
		"count": len(tree),
	})
}

// GetBatchesByLocation returns all medicine batches stored at a specific
// physical location slot. Used when the user clicks on a cell in the rack map UI.
//
// GET /api/locations/:id/batches
func GetBatchesByLocation(c *fiber.Ctx) error {
	locationId := c.Params("id")
	if locationId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing location ID"})
	}

	batches, err := dao.DB_GetBatchesByLocation(locationId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch batches for location: " + err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"locationId": locationId,
		"data":       batches,
		"count":      len(batches),
	})
}

