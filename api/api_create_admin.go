package api

import (
	"context"
	"lawyerSL-Backend/auth"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/gofiber/fiber/v2"
)

// POST /register/admin
// Creates an admin user record in the "admin_users" collection.
// Body: { name, email, password, phoneNumber }
// firebaseUid is accepted but optional (legacy compat).
func CreateAdminUser(c *fiber.Ctx) error {
	var req dto.RegisterUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name and email are required",
		})
	}

	if req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password is required",
		})
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	userID, err := dao.GenerateId(context.Background(), "admin_users", "AD")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate admin ID",
		})
	}

	admin := dto.AdminUser{
		UserID:             userID,
		FirebaseUID:        req.FirebaseUID, // kept for backward-compat, may be empty
		Name:               req.Name,
		Email:              req.Email,
		PasswordHash:       passwordHash,
		PhoneNumber:        req.PhoneNumber,
		Role:               dto.RoleAdmin,
		BranchId:           req.BranchId,
		Status:             dto.StatusActive,
		MustChangePassword: false,
		CreatedAt:          time.Now(),
	}

	if err := dao.DB_CreateAdminUser(admin); err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Never return the admin struct directly — it contains the PasswordHash (json:"-" handles this,
	// but explicit is safer)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"userId":   admin.UserID,
		"name":     admin.Name,
		"email":    admin.Email,
		"role":     admin.Role,
		"branchId": admin.BranchId,
		"status":   admin.Status,
	})
}
