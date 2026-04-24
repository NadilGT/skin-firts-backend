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

	dateStr := c.Query("date")
	if dateStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Date is required",
		})
	}

	date, err := parseFlexibleDate(dateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date format",
		})
	}

	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	runningNum, err := dao.DB_GetRunningAppointment(doctorID, date, branchId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No running appointment found for this doctor",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"running_appointment_number": runningNum,
	})
}
