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

	// For multi-branch users, determine the active branch:
	// 1. Check X-Branch-Id header
	// 2. Fallback to the first branch in the branchIds array
	branchIds, ok := c.Locals("branchIds").([]string)
	if !ok || len(branchIds) == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "No branches assigned to this user. Contact your administrator.",
		})
	}

	activeBranchId := c.Get("X-Branch-Id")
	if activeBranchId == "" {
		activeBranchId = branchIds[0] // Default to the first branch
	}

	// Validate that the requested activeBranchId is in the user's branchIds array
	isValidBranch := false
	for _, id := range branchIds {
		if id == activeBranchId {
			isValidBranch = true
			break
		}
	}

	if !isValidBranch {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You do not have access to the requested branch.",
		})
	}

	c.Locals("branchId", activeBranchId) // for downstream handlers
	c.Locals("branchFilter", bson.M{"branchId": activeBranchId})
	c.Locals("effectiveBranchId", activeBranchId)
	return c.Next()
}
