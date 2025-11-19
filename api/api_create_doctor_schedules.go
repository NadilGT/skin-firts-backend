package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/gofiber/fiber/v2"
)

// CreateDoctorSchedule creates or updates a doctor's schedule for a specific date
// POST /api/doctor-schedule
// Body: { "doctorName": "Dr. Smith", "date": "2025-11-20T00:00:00Z", "timeSlots": ["09:00 AM", "10:00 AM"] }
func CreateDoctorSchedule(c *fiber.Ctx) error {
	var req dto.CreateDoctorScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.DoctorName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Doctor name is required",
		})
	}

	if len(req.TimeSlots) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one time slot is required",
		})
	}

	schedule := dto.DoctorScheduleModel{
		DoctorName: req.DoctorName,
		Date:       req.Date,
		TimeSlots:  req.TimeSlots,
		UpdatedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}

	if err := dao.DB_CreateOrUpdateDoctorSchedule(schedule); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create doctor schedule",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Schedule created successfully",
		"schedule": schedule,
	})
}

// GetDoctorSchedule retrieves all schedules for a specific doctor
// GET /api/doctor-schedule?doctorName=Dr. Smith
func GetDoctorSchedule(c *fiber.Ctx) error {
	doctorName := c.Query("doctorName")
	if doctorName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Doctor name is required",
		})
	}

	schedules, err := dao.DB_GetDoctorSchedule(doctorName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch doctor schedule",
		})
	}

	// Convert to the response format that Flutter expects
	// Map of date string -> time slots
	scheduleMap := make(map[string][]string)
	for _, schedule := range schedules {
		dateKey := schedule.Date.Format("2006-01-02") // YYYY-MM-DD format
		scheduleMap[dateKey] = schedule.TimeSlots
	}

	response := dto.DoctorScheduleResponse{
		DoctorName: doctorName,
		Schedules:  scheduleMap,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetDoctorScheduleByDateRange retrieves schedules within a date range
// GET /api/doctor-schedule/range?doctorName=Dr. Smith&startDate=2025-11-01&endDate=2025-11-30
func GetDoctorScheduleByDateRange(c *fiber.Ctx) error {
	doctorName := c.Query("doctorName")
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")

	if doctorName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Doctor name is required",
		})
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid start date format. Use YYYY-MM-DD",
		})
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid end date format. Use YYYY-MM-DD",
		})
	}

	schedules, err := dao.DB_GetDoctorScheduleByDateRange(doctorName, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch doctor schedule",
		})
	}

	// Convert to response format
	scheduleMap := make(map[string][]string)
	for _, schedule := range schedules {
		dateKey := schedule.Date.Format("2006-01-02")
		scheduleMap[dateKey] = schedule.TimeSlots
	}

	response := dto.DoctorScheduleResponse{
		DoctorName: doctorName,
		Schedules:  scheduleMap,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// DeleteDoctorSchedule deletes a specific schedule entry
// DELETE /api/doctor-schedule?doctorName=Dr. Smith&date=2025-11-20
func DeleteDoctorSchedule(c *fiber.Ctx) error {
	doctorName := c.Query("doctorName")
	dateStr := c.Query("date")

	if doctorName == "" || dateStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Doctor name and date are required",
		})
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date format. Use YYYY-MM-DD",
		})
	}

	if err := dao.DB_DeleteDoctorSchedule(doctorName, date); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete schedule",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Schedule deleted successfully",
	})
}