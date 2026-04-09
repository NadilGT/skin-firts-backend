package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"

	"github.com/gofiber/fiber/v2"
)

func UpdateDoctorInfoByDoctorId(c *fiber.Ctx) error {
	doctorID := c.Query("doctor_id")
	if doctorID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Doctor ID parameter is required",
		})
	}

	var info dto.DoctorInfoModel
	if err := c.BodyParser(&info); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Verify focus mapping exists if it's being updated
	if info.Focus != "" {
		exists, err := dao.DB_CheckFocusExists(info.Focus)
		if err != nil || !exists {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "A valid focus must be attached and present inside focus configurations.",
			})
		}
	}

	err := dao.DB_UpdateDoctorInfoByDoctorId(doctorID, info)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update doctor info",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Doctor info updated successfully",
	})
}
