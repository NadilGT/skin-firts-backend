package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"

	"github.com/gofiber/fiber/v2"
)

func CreateDoctor(c *fiber.Ctx) error {
	var doctor dto.Doctor
	if err := c.BodyParser(&doctor); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":"Invalid request body",
		})
	}

	if err := dao.DB_CreateDoctor(doctor); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":"Failed to create employee",
		})
	}
	return c.Status(200).JSON(doctor)
}