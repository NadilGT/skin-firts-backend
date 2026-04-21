package api

import (
	"github.com/gofiber/fiber/v2"
)

// GetMyProfile returns the current user's identity from their JWT claims.
// Frontend can call this after login to get branchId and roles.
//
// GET /auth/me
func GetMyProfile(c *fiber.Ctx) error {
	uid, _ := c.Locals("uid").(string)
	email, _ := c.Locals("email").(string)
	branchId, _ := c.Locals("branchId").(string)
	role, _ := c.Locals("role").(string)
	roles, _ := c.Locals("roles").([]string)

	isSuperAdmin := false
	for _, r := range roles {
		if r == "super_admin" {
			isSuperAdmin = true
			break
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"uid":          uid,
		"email":        email,
		"branchId":     branchId,
		"role":         role,
		"roles":        roles,
		"isSuperAdmin": isSuperAdmin,
	})
}
