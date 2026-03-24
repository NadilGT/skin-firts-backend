package api

import (
	"lawyerSL-Backend/dao"

	"github.com/gofiber/fiber/v2"
)

// GET /role/admin?firebaseUid=xxx
// Looks up the role of a user in the admin_users collection.
// Used by the admin portal after Firebase login to confirm admin access.
func FindAdminRole(c *fiber.Ctx) error {
	firebaseUID := c.Query("firebaseUid")
	if firebaseUID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "firebaseUid query param is required",
		})
	}

	role, found, err := dao.DB_FindAdminRole(firebaseUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to look up role",
		})
	}
	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No admin found with this Firebase UID",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"firebaseUid": firebaseUID,
		"role":        role,
	})
}

// GET /role/mobile?firebaseUid=xxx
// Looks up the role of a user across patients and doctor_users collections.
// Used by the mobile app after Firebase login to determine which screen to show.
func FindMobileUserRole(c *fiber.Ctx) error {
	firebaseUID := c.Query("firebaseUid")
	if firebaseUID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "firebaseUid query param is required",
		})
	}

	role, found, err := dao.DB_FindMobileUserRole(firebaseUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to look up role",
		})
	}
	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No patient or doctor found with this Firebase UID",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"firebaseUid": firebaseUID,
		"role":        role,
	})
}
