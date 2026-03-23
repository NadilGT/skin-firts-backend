package api

import (
	"context"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateMedicine(c *fiber.Ctx) error {
	var medicine dto.MedicineModel
	if err := c.BodyParser(&medicine); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// Validation
	if medicine.Name == "" || medicine.Manufacturer == "" || medicine.Category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name, manufacturer, and category are required",
		})
	}

	id, err := dao.GenerateId(context.Background(), "medicines", "MED")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	medicine.MedicineId = id
	medicine.CreatedAt = time.Now()
	medicine.UpdatedAt = time.Now()
	if medicine.Status == "" {
		medicine.Status = "Active"
	}

	if err := dao.DB_CreateMedicine(medicine); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create medicine: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Medicine created successfully",
		"data":    medicine,
	})
}

func SearchMedicines(c *fiber.Ctx) error {
	var query dto.SearchMedicineQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid query parameters",
		})
	}

	// Set default pagination
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 20
	}

	medicines, total, err := dao.DB_SearchMedicines(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search medicines",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": medicines,
		"pagination": fiber.Map{
			"page":       query.Page,
			"limit":      query.Limit,
			"total":      total,
			"totalPages": (total + int64(query.Limit) - 1) / int64(query.Limit),
		},
	})
}

func GetMedicineByID(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid medicine ID",
		})
	}

	medicine, err := dao.DB_GetMedicineByID(objectID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Medicine not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": medicine,
	})
}

func UpdateMedicine(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid medicine ID",
		})
	}

	var medicine dto.MedicineModel
	if err := c.BodyParser(&medicine); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	medicine.ID = objectID
	medicine.UpdatedAt = time.Now()

	if err := dao.DB_UpdateMedicine(medicine); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update medicine",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Medicine updated successfully",
		"data":    medicine,
	})
}

func DeleteMedicine(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid medicine ID",
		})
	}

	if err := dao.DB_DeleteMedicine(objectID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete medicine",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Medicine deleted successfully",
	})
}

func GetLowStockMedicines(c *fiber.Ctx) error {
	medicines, err := dao.DB_GetLowStockMedicines()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch low stock medicines",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": medicines,
	})
}

// Debug endpoint to check what data is being sent
func DebugMedicineRequest(c *fiber.Ctx) error {
	body := c.Body()
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"receivedBody": string(body),
		"contentType":  c.Get("Content-Type"),
	})
}

// ========== BATCH APIS ==========

func CreateMedicineBatch(c *fiber.Ctx) error {
	var batch dto.MedicineBatchModel
	if err := c.BodyParser(&batch); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	batch.ID = primitive.NewObjectID()
	batch.CreatedAt = time.Now()
	
	if batch.Status == "" {
		batch.Status = "ACTIVE"
	}

	if err := dao.DB_CreateMedicineBatch(batch); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create medicine batch: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Medicine batch created successfully",
		"data":    batch,
	})
}

func GetBatchesByMedicineID(c *fiber.Ctx) error {
	id := c.Query("medicineId")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing medicineId format",
		})
	}

	batches, err := dao.DB_GetBatchesByMedicineID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve batches",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": batches,
	})
}

func GetAvailableBatchesFEFO(c *fiber.Ctx) error {
	id := c.Query("medicineId")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing medicineId format",
		})
	}

	batches, err := dao.DB_GetAvailableBatchesFEFO(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve available batches",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": batches,
	})
}

func DeductStockFEFO(c *fiber.Ctx) error {
	var req dto.DeductStockRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Quantity must be greater than zero",
		})
	}

	if req.MedicineID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing medicine ID",
		})
	}

	billItems, err := dao.DB_DeductStockFEFO(req.MedicineID, req.Quantity)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to deduct stock: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":   "Stock deducted successfully",
		"billItems": billItems,
	})
}