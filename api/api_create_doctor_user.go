package api

import (
	"context"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/gofiber/fiber/v2"
)

// POST /register/doctor-user
// Admin-only: creates the auth record for a doctor in the "doctor_users" collection.
// Body: { firebaseUid, name, email, phoneNumber }
// Generates a DOC-xxx userID.
func CreateDoctorUserAccount(c *fiber.Ctx) error {
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

	userID, err := dao.GenerateId(context.Background(), "doctor_users", "DOC")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate doctor user ID",
		})
	}

	doctorUser := dto.DoctorUser{
		UserID:      userID,
		FirebaseUID: req.FirebaseUID,
		Name:        req.Name,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Role:        dto.RoleDoctor,
		CreatedAt:   time.Now(),
	}

	if err := dao.DB_CreateDoctorUser(doctorUser); err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(doctorUser)
}
