package api

import (
	"context"
	"lawyerSL-Backend/auth"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

func CreateBranch(c *fiber.Ctx) error {
	var branch dto.BranchModel
	if err := c.BodyParser(&branch); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body: " + err.Error()})
	}
	if branch.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Branch name is required"})
	}

	id, err := dao.GenerateId(context.Background(), "branches", "BRN")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	branch.BranchId = id
	now := time.Now()
	branch.CreatedAt = &now
	branch.UpdatedAt = &now
	if branch.Status == "" {
		branch.Status = "ACTIVE"
	}

	if err := dao.DB_CreateBranch(branch); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create branch: " + err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Branch created successfully", "data": branch})
}

func GetAllBranches(c *fiber.Ctx) error {
	status := c.Query("status")
	branches, err := dao.DB_SearchBranches(status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch branches"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": branches})
}

func GetBranchByID(c *fiber.Ctx) error {
	id := c.Query("id")
	branch, err := dao.DB_GetBranchByBranchId(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Branch not found"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": branch})
}

func UpdateBranch(c *fiber.Ctx) error {
	id := c.Query("id")
	var branch dto.BranchModel
	if err := c.BodyParser(&branch); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	now := time.Now()
	branch.UpdatedAt = &now
	if err := dao.DB_UpdateBranch(id, branch); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update branch"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Branch updated successfully"})
}

func DeleteBranch(c *fiber.Ctx) error {
	id := c.Query("id")
	if err := dao.DB_DeleteBranch(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete branch"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Branch deleted successfully"})
}

// ──────────────────────────────────────────────
//  Branch Context (Role-Aware)
// ──────────────────────────────────────────────

func GetBranchContext(c *fiber.Ctx) error {
	role, _ := c.Locals("role").(string)
	jwtBranchIds, _ := c.Locals("branchIds").([]string)
	email, _ := c.Locals("email").(string)

	// Fetch live user data to get the most up-to-date branchIds from the database locally
	if email != "" {
		if liveUser, err := auth.FindUserByEmail(email); err == nil {
			if len(liveUser.BranchIds) > 0 {
				jwtBranchIds = liveUser.BranchIds
			}
			if liveUser.Role != "" {
				role = liveUser.Role
			}
		}
	}

	isSuperAdmin := false
	if roles, ok := c.Locals("roles").([]string); ok {
		for _, r := range roles {
			if r == "super_admin" || r == "SUPER_ADMIN" {
				isSuperAdmin = true
				break
			}
		}
	}

	if isSuperAdmin {
		role = "SUPER_ADMIN"
	} else if role == "" {
		role = "STAFF"
	}

	type BranchContextResponse struct {
		Role            string            `json:"role"`
		Branches        []dto.BranchModel `json:"branches"`
		DefaultBranchId string            `json:"defaultBranchId"`
		CanSelectBranch bool              `json:"canSelectBranch"`
		CurrentBranch   *dto.BranchModel  `json:"currentBranch"`
	}

	response := BranchContextResponse{
		Role:            role,
		Branches:        []dto.BranchModel{},
		CanSelectBranch: isSuperAdmin || len(jwtBranchIds) > 1, // Let users with >1 branch select
	}

	if isSuperAdmin {
		branches, err := dao.DB_SearchBranches("ACTIVE")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch branches"})
		}
		if len(branches) == 0 {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "No ACTIVE branches found in the system"})
		}
		response.Branches = branches

		var defaultBranch *dto.BranchModel
		for i := range branches {
			if branches[i].IsMainBranch {
				defaultBranch = &branches[i]
				break
			}
		}
		if defaultBranch == nil {
			defaultBranch = &branches[0]
		}

		response.DefaultBranchId = defaultBranch.BranchId
		response.CurrentBranch = defaultBranch
	} else {
		if len(jwtBranchIds) == 0 {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "User has no assigned branches"})
		}
		
		for _, bId := range jwtBranchIds {
			branch, err := dao.DB_GetBranchByBranchId(bId)
			if err == nil && branch != nil && branch.Status == "ACTIVE" {
				response.Branches = append(response.Branches, *branch)
			}
		}

		if len(response.Branches) == 0 {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "User has no ACTIVE branches"})
		}
		
		// Determine which branch is active right now
		activeBranchId := c.Get("X-Branch-Id")
		if activeBranchId == "" {
			activeBranchId = response.Branches[0].BranchId
		}
		
		response.DefaultBranchId = response.Branches[0].BranchId
		for i := range response.Branches {
			if response.Branches[i].BranchId == activeBranchId {
				response.CurrentBranch = &response.Branches[i]
				break
			}
		}
		if response.CurrentBranch == nil {
			response.CurrentBranch = &response.Branches[0]
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

