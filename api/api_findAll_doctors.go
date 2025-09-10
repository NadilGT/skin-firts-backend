package api

import (
	"lawyerSL-Backend/dao"

	"github.com/gofiber/fiber/v2"
)

func FindAllDoctors(c *fiber.Ctx) error {
	returnValue, err := dao.DB_FindAllDoctors()
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusAccepted).JSON(returnValue)
}