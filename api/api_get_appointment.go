package api

import (
	"lawyerSL-Backend/dao"

	"github.com/gofiber/fiber/v2"
)

func GetAppointmentByID(c *fiber.Ctx) error {
	appointmentID := c.Query("appointmentId")
	if appointmentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Appointment ID is required",
		})
	}

	appointment, err := dao.DB_GetAppointmentByAppointmentID(appointmentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Appointment not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(appointment)
}
