package api

import (
	"lawyerSL-Backend/dao"

	"github.com/gofiber/fiber/v2"
)

// SaveFCMTokenRequest is the request body for POST /api/users/save-token
type SaveFCMTokenRequest struct {
	FirebaseUID string `json:"userId"`   // Firebase UID sent from Flutter
	FcmToken    string `json:"fcmToken"` // FCM device token
}

// SaveFCMToken handles POST /api/users/save-token
// It stores (or updates) the patient's FCM device token in the patients collection.
func SaveFCMToken(c *fiber.Ctx) error {
	var req SaveFCMTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.FirebaseUID == "" || req.FcmToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userId and fcmToken are required",
		})
	}

	if err := dao.DB_SavePatientFCMToken(req.FirebaseUID, req.FcmToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save FCM token: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "FCM token saved successfully",
	})
}
