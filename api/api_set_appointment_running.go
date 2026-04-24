package api

import (
	"lawyerSL-Backend/dao"
	"time"

	"github.com/gofiber/fiber/v2"

)

func SetAppointmentRunning(c *fiber.Ctx) error {
	idParam := c.Query("appointmentId")
	if idParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Appointment ID is required",
		})
	}

	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := dao.DB_UpdateAppointmentStatus(idParam, "running", branchId); err != nil {
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
