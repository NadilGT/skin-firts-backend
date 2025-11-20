package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func UpdateAppointmentStatus(c *fiber.Ctx) error {
	idParam := c.Params("id")

	appointmentID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid appointment ID",
		})
	}

	var req dto.UpdateAppointmentStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update in DB
	if err := dao.DB_UpdateAppointmentStatus(appointmentID, req.Status); err != nil {
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
