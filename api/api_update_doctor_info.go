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

	targetBranchId := ""

	// Check if the user is a super admin
	roles, _ := c.Locals("roles").([]string)
	isSuper := false
	for _, r := range roles {
		if r == "super_admin" {
			isSuper = true
			break
		}
	}

	effectiveBranchId, _ := c.Locals("effectiveBranchId").(string)

	if isSuper {
		// Super Admin can provide arbitrary branchId via body
		var req DoctorBranchAssignRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}
		if req.BranchId == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "BranchId is required in body for super_admin"})
		}
		targetBranchId = req.BranchId
	} else {
		// Normal branch admin can only assign the doctor to their own branch
		if effectiveBranchId == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You do not have a branch assigned"})
		}
		targetBranchId = effectiveBranchId
	}

	err := dao.DB_AddDoctorToBranch(doctorID, targetBranchId)
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

	targetBranchId := ""

	roles, _ := c.Locals("roles").([]string)
	isSuper := false
	for _, r := range roles {
		if r == "super_admin" {
			isSuper = true
			break
		}
	}

	effectiveBranchId, _ := c.Locals("effectiveBranchId").(string)

	if isSuper {
		var req DoctorBranchAssignRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}
		if req.BranchId == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "BranchId is required in body for super_admin"})
		}
		targetBranchId = req.BranchId
	} else {
		if effectiveBranchId == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You do not have a branch assigned"})
		}
		targetBranchId = effectiveBranchId
	}

	err := dao.DB_RemoveDoctorFromBranch(doctorID, targetBranchId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to remove doctor from branch: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Doctor successfully removed from branch",
		"branchId": targetBranchId,
	})
}
