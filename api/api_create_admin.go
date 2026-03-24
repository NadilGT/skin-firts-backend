package api

import (
	"context"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/gofiber/fiber/v2"
)

// POST /register/admin
// Admin-only: creates an admin user record in the "admin_users" collection.
// Body: { firebaseUid, name, email, phoneNumber }
// Generates an AD-xxx userID.
func CreateAdminUser(c *fiber.Ctx) error {
	var req dto.RegisterUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.FirebaseUID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "firebaseUid is required",
		})
	}
	if req.Name == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name and email are required",
		})
	}

	userID, err := dao.GenerateId(context.Background(), "admin_users", "AD")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate admin ID",
		})
	}

	admin := dto.AdminUser{
		UserID:      userID,
		FirebaseUID: req.FirebaseUID,
		Name:        req.Name,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Role:        dto.RoleAdmin,
		CreatedAt:   time.Now(),
	}

	if err := dao.DB_CreateAdminUser(admin); err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(admin)
}
