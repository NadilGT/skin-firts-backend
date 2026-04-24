package api

import (
	"lawyerSL-Backend/dao"
	"log"

	"github.com/gofiber/fiber/v2"
)


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

type RequestUser struct {
	Role     string
	BranchId string
	UserId   string
}

// GetUserFromContext extracts the user profile from the Fiber context.
func GetUserFromContext(c *fiber.Ctx) RequestUser {
	role := "PATIENT" // default fallback
	if IsSuperAdmin(c) {
		role = "SUPER_ADMIN"
	} else {
		tokenRole, _ := c.Locals("role").(string)
		if tokenRole != "" {
			switch tokenRole {
			case "admin", "ADMIN":
				role = "ADMIN"
			case "staff", "STAFF":
				role = "STAFF"
			case "patient", "PATIENT":
				role = "PATIENT"
			}
		}
	}

	jwtBranchId, _ := c.Locals("branchId").(string)
	uid, _ := c.Locals("uid").(string)

	return RequestUser{
		Role:     role,
		BranchId: jwtBranchId,
		UserId:   uid,
	}
}

// ResolveBranchId centralizes the branch enforcement rule.
func ResolveBranchId(c *fiber.Ctx, bodyBranchId string) (string, error) {
	user := GetUserFromContext(c)

	switch user.Role {
	case "ADMIN", "STAFF":
		// Always force token branch
		if bodyBranchId != "" && bodyBranchId != user.BranchId {
			log.Printf(
				"[SECURITY WARNING] Branch spoof attempt | user=%s role=%s body=%s token=%s\n",
				user.UserId,
				user.Role,
				bodyBranchId,
				user.BranchId,
			)
		}
		return user.BranchId, nil

	case "SUPER_ADMIN", "PATIENT":
		if bodyBranchId == "" {
			return "", fiber.NewError(fiber.StatusBadRequest, "branchId is required")
		}

		branch, err := dao.DB_GetBranchByBranchId(bodyBranchId)
		if err != nil || branch == nil || branch.Status != "ACTIVE" {
			return "", fiber.NewError(fiber.StatusBadRequest, "Invalid or inactive branchId")
		}

		return bodyBranchId, nil

	default:
		return "", fiber.NewError(fiber.StatusForbidden, "Unauthorized role")
	}
}
