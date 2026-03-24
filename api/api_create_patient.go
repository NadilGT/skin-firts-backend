package api

import (
	"context"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/gofiber/fiber/v2"
)

// POST /register/patient
// Body: { firebaseUid, name, email, phoneNumber }
// Generates a PAT-xxx userID and stores the patient in the "patients" collection.
func CreatePatientUser(c *fiber.Ctx) error {
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

	userID, err := dao.GenerateId(context.Background(), "patients", "PAT")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate patient ID",
		})
	}

	patient := dto.PatientUser{
		UserID:      userID,
		FirebaseUID: req.FirebaseUID,
		Name:        req.Name,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Role:        dto.RolePatient,
		CreatedAt:   time.Now(),
	}

	if err := dao.DB_CreatePatient(patient); err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(patient)
}
