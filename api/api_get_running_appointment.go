package api

import (
	"lawyerSL-Backend/dao"

	"github.com/gofiber/fiber/v2"
)

func GetRunningAppointmentNumber(c *fiber.Ctx) error {
	doctorID := c.Query("doctorId")
	if doctorID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Doctor ID is required",
		})
	}

	runningNum, err := dao.DB_GetRunningAppointment(doctorID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No running appointment found for this doctor",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"running_appointment_number": runningNum,
	})
}
