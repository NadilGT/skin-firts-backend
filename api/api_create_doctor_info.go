package api

import (
	"context"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/utils"

	"github.com/gofiber/fiber/v2"
)

func CreateDoctorInfo(c *fiber.Ctx) error {
	var info dto.DoctorInfoModel
	id, err := dao.GenerateId(context.Background(), "doctor_info", "DOC")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	info.DoctorID = id
	if err := c.BodyParser(&info); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Verify focus mapping exists before binding it to the doctor
	focusToCheck := info.FocusId
	if focusToCheck == "" {
		focusToCheck = info.Focus
	}

	exists, err := dao.DB_CheckFocusExists(focusToCheck)
	if err != nil || !exists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "A valid focus (FocusID or Name) must be attached and present inside focus configurations.",
		})
	}

	if err := dao.DB_CreateDoctorInfo(info); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create doctor info",
		})
	}

	return c.Status(fiber.StatusOK).JSON(info)
}