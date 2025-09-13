package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"

	"github.com/gofiber/fiber/v2"
)

func CreateDoctorInfo(c *fiber.Ctx) error {
	var info dto.DoctorInfoModel
	if err := c.BodyParser(&info); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := dao.DB_CreateDoctorInfo(info); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create doctor info",
		})
	}

	return c.Status(fiber.StatusOK).JSON(info)
}