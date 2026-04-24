package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"math"
	"strconv"
	"time"

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
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	appointments, total, err := dao.DB_FindAllAppointments(branchId, page, limit)
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
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	appointments, total, err := dao.DB_FindAppointmentsByDoctorID(doctorID, branchId, page, limit)
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
// GET /findAll/appointments/doctor/ordered?doctorId=DOC-001&date=2025-11-18&page=1&limit=10
// Returns appointments for a doctor on a specific date, sorted by appointmentNumber ascending (1, 2, 3 …)
func GetAppointmentsByDoctorIDSortedByNumber(c *fiber.Ctx) error {
	doctorID := c.Query("doctorId")
	dateStr := c.Query("date")

	if doctorID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Doctor ID is required",
		})
	}

	if dateStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Date is required (YYYY-MM-DD format)",
		})
	}

	// Parse date (expecting YYYY-MM-DD)
	appointmentDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date format. Use YYYY-MM-DD",
		})
	}

	page, limit := parsePagination(c)
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	appointments, total, err := dao.DB_FindAppointmentsByDoctorIDSortedByNumber(doctorID, appointmentDate, branchId, page, limit)
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

	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	page, limit := parsePagination(c)

	appointments, total, err := dao.DB_FindAppointmentsByPatientID(patientID, branchId, page, limit)
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

// GET /findAll/appointments/doctor/detailed?doctorId=DOC-001&date=2025-11-18&status=pending&page=1&limit=10
// Returns appointments for a doctor on a specific date with a specific status, sorted by appointmentNumber ascending.
func GetAppointmentsByDoctorDateStatus(c *fiber.Ctx) error {
	doctorID := c.Query("doctorId")
	dateStr := c.Query("date")
	status := c.Query("status")

	if doctorID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Doctor ID is required",
		})
	}

	if dateStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Date is required (YYYY-MM-DD format)",
		})
	}

	if status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Status is required",
		})
	}

	// Parse date (expecting YYYY-MM-DD)
	appointmentDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date format. Use YYYY-MM-DD",
		})
	}

	page, limit := parsePagination(c)
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	appointments, total, err := dao.DB_FindAppointmentsByDoctorDateStatus(doctorID, appointmentDate, status, branchId, page, limit)
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
