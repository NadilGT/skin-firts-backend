package api

import (
	"lawyerSL-Backend/dao"

	"github.com/gofiber/fiber/v2"
)

func GetFavoriteDoctors(c *fiber.Ctx)error{
	doctors, err := dao.DB_GetFavoriteDoctors()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(doctors)
}