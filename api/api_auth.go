package api

import (
	"github.com/gofiber/fiber/v2"
)

// GetMyProfile returns the current user's identity from their JWT claims.
// Frontend can call this after login to get branchId, role, and status.
//
// GET /auth/me  (also kept at this handler for legacy routing)
//
// Locals read (set by auth.JWTMiddleware):
//   - userId   (also aliased as uid for backward-compat)
//   - email
//   - branchId
//   - role
//   - roles    []string
func GetMyProfile(c *fiber.Ctx) error {
	userId, _ := c.Locals("userId").(string)
	uid, _ := c.Locals("uid").(string) // backward-compat alias
	if userId == "" {
		userId = uid
	}

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
		"userId":       userId,
		"uid":          userId, // backward-compat field
		"email":        email,
		"branchId":     branchId,
		"role":         role,
		"roles":        roles,
		"isSuperAdmin": isSuperAdmin,
	})
}
