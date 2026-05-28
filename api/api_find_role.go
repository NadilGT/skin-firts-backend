package api

import (
	"lawyerSL-Backend/dao"

	"github.com/gofiber/fiber/v2"
)

// GET /role/admin?email=xxx  (or legacy ?firebaseUid=xxx)
// Looks up the role of a user in the admin_users collection.
// Supports both email (new) and firebaseUid (legacy) for backward-compat.
func FindAdminRole(c *fiber.Ctx) error {
	// Support both new (email) and legacy (firebaseUid) identifiers
	identifier := c.Query("email")
	if identifier == "" {
		identifier = c.Query("firebaseUid") // legacy fallback
	}
	if identifier == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email query param is required (or firebaseUid for legacy clients)",
		})
	}

	role, found, err := dao.DB_FindAdminRole(identifier)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to look up role",
		})
	}
	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No admin found with this identifier",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"identifier": identifier,
		"role":       role,
	})
}

// GET /role/mobile?email=xxx  (or legacy ?firebaseUid=xxx)
// Looks up the role of a user across patients and doctor_users collections.
func FindMobileUserRole(c *fiber.Ctx) error {
	identifier := c.Query("email")
	if identifier == "" {
		identifier = c.Query("firebaseUid") // legacy fallback
	}
	if identifier == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email query param is required (or firebaseUid for legacy clients)",
		})
	}

	role, found, err := dao.DB_FindMobileUserRole(identifier)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to look up role",
		})
	}
	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No patient or doctor found with this identifier",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"identifier": identifier,
		"role":       role,
	})
}
