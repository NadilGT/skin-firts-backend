package api

import (
	"lawyerSL-Backend/dao"

	"github.com/gofiber/fiber/v2"
)

func GetNextAppointmentNumber(c *fiber.Ctx) error {
	doctorID := c.Query("doctorId")
	if doctorID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Doctor ID is required",
		})
	}

	nextNum, err := dao.DB_GetNextAppointmentNumber(doctorID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get next appointment number",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"next_appointment_number": nextNum,
	})
}
