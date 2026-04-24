package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"

	"github.com/gofiber/fiber/v2"
)

func FindAllDoctors(c *fiber.Ctx) error {
	returnValue, err := dao.DB_FindAllDoctors()
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusAccepted).JSON(returnValue)
}

func GetDoctorsByFocus(c *fiber.Ctx) error {
	focusId := c.Query("focusId")
	if focusId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query parameter 'focusId' is required",
		})
	}

	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	doctors, err := dao.DB_FindDoctorsByFocus(focusId, branchId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch doctors by focus",
		})
	}

	if doctors == nil || len(*doctors) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "No doctors found with this focus",
			"doctors": []dto.DoctorInfoModel{},
		})
	}

	return c.Status(fiber.StatusOK).JSON(doctors)
}