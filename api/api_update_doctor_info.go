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

// Request struct for branch assignment/removal
type DoctorBranchAssignRequest struct {
	BranchId string `json:"branchId"`
}

// AssignDoctorToBranch appends a branchId to a doctor's BranchIds array.
func AssignDoctorToBranch(c *fiber.Ctx) error {
	doctorID := c.Query("doctor_id")
	if doctorID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Doctor ID parameter is required"})
	}

	var req DoctorBranchAssignRequest
	_ = c.BodyParser(&req)

	targetBranchId, err := ResolveBranchId(c, req.BranchId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	err = dao.DB_AddDoctorToBranch(doctorID, targetBranchId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to assign doctor to branch: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Doctor successfully assigned to branch",
		"branchId": targetBranchId,
	})
}

// RemoveDoctorFromBranch removes a branchId from a doctor's BranchIds array.
func RemoveDoctorFromBranch(c *fiber.Ctx) error {
	doctorID := c.Query("doctor_id")
	if doctorID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Doctor ID parameter is required"})
	}

	var req DoctorBranchAssignRequest
	_ = c.BodyParser(&req)

	targetBranchId, err := ResolveBranchId(c, req.BranchId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	err = dao.DB_RemoveDoctorFromBranch(doctorID, targetBranchId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to remove doctor from branch: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Doctor successfully removed from branch",
		"branchId": targetBranchId,
	})
}
