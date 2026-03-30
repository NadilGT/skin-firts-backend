package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/gofiber/fiber/v2"

)

func UpdateAppointmentStatus(c *fiber.Ctx) error {
	idParam := c.Query("appointmentId")

	if idParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Appointment ID is required",
		})
	}

	var req dto.UpdateAppointmentStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update in DB
	if err := dao.DB_UpdateAppointmentStatus(idParam, req.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update appointment status",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Appointment status updated successfully",
		"status":  req.Status,
		"updatedAt": time.Now(),
	})
}
