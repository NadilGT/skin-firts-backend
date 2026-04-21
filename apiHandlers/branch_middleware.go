package apiHandlers

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// BranchMiddleware injects a MongoDB branch filter into the request context.
//
// Usage:
//   - super_admin → empty filter (sees all branches)
//   - all other roles → filter by their branchId
//
// Call AFTER ValidateToken so that c.Locals("branchId") and c.Locals("roles") are set.
func BranchMiddleware(c *fiber.Ctx) error {
	roles, _ := c.Locals("roles").([]string)

	// super_admin bypasses branch isolation
	for _, r := range roles {
		if r == "super_admin" {
			c.Locals("branchFilter", bson.M{})
			c.Locals("effectiveBranchId", "") // empty = no restriction
			return c.Next()
		}
	}

	branchId, _ := c.Locals("branchId").(string)
	if branchId == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Branch not assigned to this user. Contact your administrator.",
		})
	}

	c.Locals("branchFilter", bson.M{"branchId": branchId})
	c.Locals("effectiveBranchId", branchId)
	return c.Next()
}
