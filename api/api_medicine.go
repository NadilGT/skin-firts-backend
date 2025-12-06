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

	if medicine.Price <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Price must be greater than 0",
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