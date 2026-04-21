package api

import "github.com/gofiber/fiber/v2"

// GetBranchId is a helper for API handlers to safely extract the effective branchId.
// Returns empty string for super_admin (they see all data).
func GetBranchId(c *fiber.Ctx) string {
	branchId, _ := c.Locals("effectiveBranchId").(string)
	return branchId
}

// IsSuperAdmin checks if the current user is a super_admin.
func IsSuperAdmin(c *fiber.Ctx) bool {
	roles, _ := c.Locals("roles").([]string)
	for _, r := range roles {
		if r == "super_admin" {
			return true
		}
	}
	return false
}
