package api

import (
	"lawyerSL-Backend/dao"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SetAppointmentRunning(c *fiber.Ctx) error {
	idParam := c.Query("id")

	appointmentID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid appointment ID",
		})
	}

	if err := dao.DB_UpdateAppointmentStatus(appointmentID, "running"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to set appointment status to running",
		})
	}

	return c.JSON(fiber.Map{
		"message":   "Appointment status updated to running successfully",
		"status":    "running",
		"updatedAt": time.Now(),
	})
}
