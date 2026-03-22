package api

import (
	"context"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"

	"github.com/gofiber/fiber/v2"
)

func CreateFocus(c *fiber.Ctx) error {
	focusName := c.Query("focus")

	if focusName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Focus query parameter is required (e.g., ?focus=Dermatology)",
		})
	}

	var req dto.FocusModel
	req.Name = focusName

	id, err := dao.GenerateId(context.Background(), "focus", "FOC")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate focus ID",
		})
	}
	req.FocusID = id

	err = dao.DB_CreateFocus(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create focus category",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Focus category inserted successfully",
		"focus":   req,
	})
}

func GetAllFocuses(c *fiber.Ctx) error {
	focuses, err := dao.DB_GetAllFocuses()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve focus categories",
		})
	}

	if len(focuses) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "No focus categories found",
			"focuses": []dto.FocusModel{},
		})
	}

	return c.Status(fiber.StatusOK).JSON(focuses)
}
