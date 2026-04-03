package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"

	"github.com/gofiber/fiber/v2"
)

// SearchDoctorInfo handles requests for searching doctor clinical profiles.
// GET /doctors/search?query=...&focus=...&special=...&page=...&limit=...
func SearchDoctorInfo(c *fiber.Ctx) error {
	var query dto.SearchDoctorInfoQuery
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

	doctors, total, err := dao.DB_SearchDoctorInfo(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search doctor info: " + err.Error(),
		})
	}

	totalPages := (total + int64(query.Limit) - 1) / int64(query.Limit)

	return c.Status(fiber.StatusOK).JSON(dto.DoctorInfoSearchResponse{
		Data:       doctors,
		Total:      total,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: int(totalPages),
	})
}
