package api

import (
	"lawyerSL-Backend/dao"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RescheduleAppointment(c *fiber.Ctx) error {
	appointmentID := c.Query("appointmentId")
	if appointmentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Appointment ID is required",
		})
	}

	var req struct {
		AppointmentDate string `json:"appointmentDate"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.AppointmentDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Appointment date is required",
		})
	}

	// Parse the new date
	newDate, err := parseFlexibleDate(req.AppointmentDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date format. Use ISO8601 or YYYY-MM-DD",
		})
	}

	// Prevent past bookings
	if newDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot reschedule appointment to a past date",
		})
	}

	// Get existing appointment to check availability for the same doctor/time
	existingAppointment, err := dao.DB_GetAppointmentByID(appointmentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Appointment not found",
		})
	}

	// Check if the doctor is available on the new date
	isAvailable, reason, err := dao.DB_CheckDoctorAvailabilityOnDate(existingAppointment.DoctorID, newDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to check doctor availability",
			"details": err.Error(),
		})
	}
	if !isAvailable {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": reason,
		})
	}

	// Update the appointment
	if err := dao.DB_RescheduleAppointment(appointmentID, newDate); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reschedule appointment",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Appointment rescheduled successfully",
	})
}