package api

import (
	"context"
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
