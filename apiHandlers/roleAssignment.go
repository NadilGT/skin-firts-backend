// apiHandlers/roleAssignment.go
// Role management — fully MongoDB-based. Firebase dependency removed.
package apiHandlers

import (
	"lawyerSL-Backend/dao"

	"github.com/gofiber/fiber/v2"
)

// ---------------------------------------------------------------------------
// Request / handler types
// ---------------------------------------------------------------------------

type RoleAssignmentRequest struct {
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	BranchIds []string `json:"branchIds"`
}

// RoleAssignmentHandler handles role management via MongoDB.
type RoleAssignmentHandler struct{}

func NewRoleAssignmentHandler(_ ...interface{}) *RoleAssignmentHandler {
	return &RoleAssignmentHandler{}
}

// ---------------------------------------------------------------------------
// POST /admin/assign-roles
// ---------------------------------------------------------------------------

// AssignRoles updates a user's role and branchIds in MongoDB.
// The user is identified by email across all 4 user collections.
func (h *RoleAssignmentHandler) AssignRoles(c *fiber.Ctx) error {
	var req RoleAssignmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" || len(req.Roles) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email and at least one role are required",
		})
	}

	// Use the first role as the primary role
	primaryRole := req.Roles[0]

	if err := dao.DB_UpdateUserRoleAndBranches(req.Email, primaryRole, req.BranchIds); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Roles assigned successfully",
		"email":    req.Email,
		"roles":    req.Roles,
		"branchIds": req.BranchIds,
		"note":     "Changes take effect on next login (new JWT issued)",
	})
}

// ---------------------------------------------------------------------------
// GET /admin/user-roles?email=xxx
// ---------------------------------------------------------------------------

// GetUserRoles returns the role and branchId for a user found by email.
func (h *RoleAssignmentHandler) GetUserRoles(c *fiber.Ctx) error {
	email := c.Query("email")
	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email query parameter is required",
		})
	}

	role, found, err := dao.DB_FindAdminRole(email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}
	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"email": email,
		"role":  role,
	})
}

// ---------------------------------------------------------------------------
// GET /admin/list-users
// ---------------------------------------------------------------------------

// ListAllUsers returns a sanitised list of all users from all collections.
// passwordHash is never included.
func (h *RoleAssignmentHandler) ListAllUsers(c *fiber.Ctx) error {
	users, err := dao.DB_ListAllUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list users: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"users": users,
		"count": len(users),
	})
}

// ---------------------------------------------------------------------------
// DELETE /admin/remove-roles?email=xxx
// ---------------------------------------------------------------------------

// RemoveRoles clears the role and branchId for a user (sets to empty string).
func (h *RoleAssignmentHandler) RemoveRoles(c *fiber.Ctx) error {
	email := c.Query("email")
	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email query parameter is required",
		})
	}

	if err := dao.DB_UpdateUserRoleAndBranches(email, "", []string{}); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Roles removed successfully",
		"email":   email,
	})
}

// ---------------------------------------------------------------------------
// PATCH /admin/user-status
// ---------------------------------------------------------------------------

type UpdateStatusRequest struct {
	Email  string `json:"email"`
	Status string `json:"status"` // ACTIVE | INACTIVE | SUSPENDED
}

// UpdateUserStatus allows admins to activate, deactivate or suspend an account.
func UpdateUserStatus(c *fiber.Ctx) error {
	var req UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	allowed := map[string]bool{"ACTIVE": true, "INACTIVE": true, "SUSPENDED": true}
	if !allowed[req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "status must be ACTIVE, INACTIVE, or SUSPENDED",
		})
	}

	if err := dao.DB_UpdateUserStatus(req.Email, req.Status); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User status updated",
		"email":   req.Email,
		"status":  req.Status,
	})
}
