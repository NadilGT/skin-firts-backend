package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// parsePagination reads ?page=&limit= from the query string.
// Defaults: page=1, limit=10. Maximum limit capped at 100.
func parsePagination(c *fiber.Ctx) (int, int) {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

// GET /findAll/appointments?page=1&limit=10
func GetAllAppointments(c *fiber.Ctx) error {
	page, limit := parsePagination(c)

	appointments, total, err := dao.DB_FindAllAppointments(page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch appointments",
		})
	}

	if appointments == nil {
		appointments = []dto.AppointmentModel{}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":       appointments,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
	})
}

// GET /findAll/appointments/doctor?doctorId=DOC-001&page=1&limit=10
func GetAppointmentsByDoctorID(c *fiber.Ctx) error {
	doctorID := c.Query("doctorId")
	if doctorID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Doctor ID is required",
		})
	}

	page, limit := parsePagination(c)

	appointments, total, err := dao.DB_FindAppointmentsByDoctorID(doctorID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch appointments for this doctor",
		})
	}

	if appointments == nil {
		appointments = []dto.AppointmentModel{}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":       appointments,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
	})
}
// GET /findAll/appointments/patient?patientId=oUTllPRCeeNiwEXKEhtTlSOZu4w1&page=1&limit=10
func GetAppointmentsByPatientID(c *fiber.Ctx) error {
	patientID := c.Query("patientId")
	if patientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Patient ID is required",
		})
	}

	page, limit := parsePagination(c)

	appointments, total, err := dao.DB_FindAppointmentsByPatientID(patientID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch appointments for this patient",
		})
	}

	if appointments == nil {
		appointments = []dto.AppointmentModel{}
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":       appointments,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
	})
}
