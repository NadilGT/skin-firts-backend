package api

import (
	"lawyerSL-Backend/dao"

	"github.com/gofiber/fiber/v2"
)

func FindDoctorInfoByName(c *fiber.Ctx) error {
	name := c.Query("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name parameter is required",
		})
	}

	info, err := dao.DB_GetDoctorInfoByName(name)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Doctor info not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(info)
}
