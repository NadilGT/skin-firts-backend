package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"

	"github.com/gofiber/fiber/v2"
)

// GetStockMovements returns paginated stock movements with optional filters.
// Supports filtering by batchId, medicineId, branchId, type, referenceId, date range.
//
// GET /stock-movements?batchId=&medicineId=&branchId=&type=&from=&to=&page=&limit=
func GetStockMovements(c *fiber.Ctx) error {
	var query dto.SearchMovementQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query parameters"})
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 50
	}
	movements, total, err := dao.DB_SearchStockMovements(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch stock movements: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": movements,
		"pagination": fiber.Map{
			"page":       query.Page,
			"limit":      query.Limit,
			"total":      total,
			"totalPages": (total + int64(query.Limit) - 1) / int64(query.Limit),
		},
	})
}

// GetMovementsByBatch returns the full ledger history for a specific batch.
//
// GET /stock-movements/batch/:batchId
func GetMovementsByBatch(c *fiber.Ctx) error {
	batchId := c.Params("batchId")
	if batchId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing batchId"})
	}
	movements, err := dao.DB_GetMovementsByBatch(batchId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch movements: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":  movements,
		"count": len(movements),
	})
}
