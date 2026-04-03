package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"

	"github.com/gofiber/fiber/v2"
)

// SearchPatients handles requests for searching patient records.
// GET /admin/search-patients?query=...&page=...&limit=...
func SearchPatients(c *fiber.Ctx) error {
	var query dto.SearchPatientQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid query parameters: " + err.Error(),
		})
	}

	// Default pagination
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	patients, total, err := dao.DB_SearchPatients(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search patients: " + err.Error(),
		})
	}

	totalPages := (total + int64(query.Limit) - 1) / int64(query.Limit)

	return c.Status(fiber.StatusOK).JSON(dto.PatientSearchResponse{
		Data:       patients,
		Total:      total,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: int(totalPages),
	})
}
