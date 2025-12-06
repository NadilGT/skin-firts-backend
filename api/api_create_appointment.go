package api

import (
	"fmt"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/dao"
	"time"

	"github.com/gofiber/fiber/v2"
)

var dateFormats = []string{
	time.RFC3339,              // 2025-11-18T10:30:00Z
	"2006-01-02",              // 2025-11-18
	"2006-01-02 15:04",        // optional
	"2006-01-02 15:04:05",
}

func parseFlexibleDate(dateStr string) (time.Time, error) {
	for _, format := range dateFormats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date")
}

func CreateAppointment(c *fiber.Ctx) error {
	var req dto.CreateAppointmentRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	fmt.Println("Incoming appointmentDate:", req.AppointmentDate)

	// Parse flexible date input
	appointmentDate, err := parseFlexibleDate(req.AppointmentDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date format. Use ISO8601 or YYYY-MM-DD",
		})
	}

	// Prevent past bookings
	if appointmentDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot book appointment in the past",
		})
	}

	// Check availability
	available, err := dao.DB_IsTimeSlotAvailable(req.DoctorID, appointmentDate, req.TimeSlot)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check time slot availability",
		})
	}

	if !available {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "This time slot is already booked",
		})
	}

	// Build the model
	appointment := dto.AppointmentModel{
		PatientID:       req.PatientID,
		PatientName:     req.PatientName,
		PatientEmail:    req.PatientEmail,
		PatientPhone:    req.PatientPhone,
		DoctorID:        req.DoctorID,
		DoctorName:      req.DoctorName,
		DoctorSpecialty: req.DoctorSpecialty,
		AppointmentDate: appointmentDate,
		TimeSlot:        req.TimeSlot,
		Status:          "pending",
		Notes:           req.Notes,
	}

	// Save to DB
	if err := dao.DB_CreateAppointment(appointment); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create appointment",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Appointment booked successfully",
		"appointment": appointment,
	})
}
