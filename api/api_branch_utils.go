package api

import (
	"lawyerSL-Backend/dao"
	"log"

	"github.com/gofiber/fiber/v2"
)

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

// EnforceBranchId ensures branch security rules are strictly followed.
// - SUPER_ADMIN: must provide a valid, ACTIVE branchId in the request body.
// - OTHERS: securely overrides any provided branchId with the one from their JWT. Logs spoof attempts.
func EnforceBranchId(targetBranchId *string, c *fiber.Ctx) error {
	role, _ := c.Locals("role").(string)
	jwtBranchId, _ := c.Locals("branchId").(string)
	uid, _ := c.Locals("uid").(string)

	isSuperAdmin := IsSuperAdmin(c)

	if isSuperAdmin {
		if targetBranchId == nil || *targetBranchId == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "branchId is mandatory in the request body for super_admin"})
		}
		branch, err := dao.DB_GetBranchByBranchId(*targetBranchId)
		if err != nil || branch == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid branchId: branch does not exist"})
		}
		if branch.Status != "ACTIVE" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid branchId: branch is not ACTIVE"})
		}
		return nil
	}

	// Non-super admin
	if targetBranchId != nil && *targetBranchId != "" && *targetBranchId != jwtBranchId {
		log.Printf("[SECURITY WARNING] Spoof Attempt: userId=%s role=%s bodyBranchId=%s tokenBranchId=%s\n", uid, role, *targetBranchId, jwtBranchId)
	}
	
	if targetBranchId != nil {
		*targetBranchId = jwtBranchId
	}
	return nil
}
