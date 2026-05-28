package api

import (
	"context"
	"lawyerSL-Backend/auth"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/gofiber/fiber/v2"
)

// POST /register/doctor-user
// Admin-only: creates the auth record for a doctor in the "doctor_users" collection.
// Body: { name, email, password, phoneNumber }
// firebaseUid is accepted but optional (legacy compat).
func CreateDoctorUserAccount(c *fiber.Ctx) error {
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

	userID, err := dao.GenerateId(context.Background(), "doctor_users", "DOC")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate doctor user ID",
		})
	}

	doctorUser := dto.DoctorUser{
		UserID:             userID,
		FirebaseUID:        req.FirebaseUID, // optional legacy field
		Name:               req.Name,
		Email:              req.Email,
		PasswordHash:       passwordHash,
		PhoneNumber:        req.PhoneNumber,
		Role:               dto.RoleDoctor,
		BranchIds:          req.BranchIds,
		Status:             dto.StatusActive,
		MustChangePassword: false,
		CreatedAt:          time.Now(),
	}

	if err := dao.DB_CreateDoctorUser(doctorUser); err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"userId":   doctorUser.UserID,
		"name":     doctorUser.Name,
		"email":    doctorUser.Email,
		"role":     doctorUser.Role,
		"branchIds": doctorUser.BranchIds,
		"status":   doctorUser.Status,
	})
}
