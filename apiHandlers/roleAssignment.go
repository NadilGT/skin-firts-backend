// apiHandlers/roleAssignment.go
package apiHandlers

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"github.com/gofiber/fiber/v2"
)

type RoleAssignmentRequest struct {
	Email string   `json:"email"`
	Roles []string `json:"roles"`
}

type RoleAssignmentHandler struct {
	firebaseApp *firebase.App
}

func NewRoleAssignmentHandler(firebaseApp *firebase.App) *RoleAssignmentHandler {
	return &RoleAssignmentHandler{
		firebaseApp: firebaseApp,
	}
}

// AssignRoles - Assign roles to a user (SUPER ADMIN ONLY - Use carefully!)
func (h *RoleAssignmentHandler) AssignRoles(c *fiber.Ctx) error {
	ctx := context.Background()

	// Get Firebase Auth client
	client, err := h.firebaseApp.Auth(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to initialize Firebase Auth",
		})
	}

	var req RoleAssignmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get user by email
	user, err := client.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Set custom claims (roles)
	claims := map[string]interface{}{
		"roles": req.Roles,
	}

	err = client.SetCustomUserClaims(ctx, user.UID, claims)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to assign roles",
		})
	}

	log.Printf("✅ Roles assigned to user %s: %v", req.Email, req.Roles)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Roles assigned successfully",
		"email":   req.Email,
		"roles":   req.Roles,
		"note":    "User must log out and log back in to get new token with roles",
	})
}

// GetUserRoles - Get current roles for a user
func (h *RoleAssignmentHandler) GetUserRoles(c *fiber.Ctx) error {
	ctx := context.Background()

	client, err := h.firebaseApp.Auth(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to initialize Firebase Auth",
		})
	}

	email := c.Query("email")
	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email query parameter required",
		})
	}

	user, err := client.GetUserByEmail(ctx, email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	roles := []string{}
	if user.CustomClaims != nil {
		if rolesInterface, ok := user.CustomClaims["roles"]; ok {
			if rolesList, ok := rolesInterface.([]interface{}); ok {
				for _, role := range rolesList {
					if roleStr, ok := role.(string); ok {
						roles = append(roles, roleStr)
					}
				}
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"email": email,
		"uid":   user.UID,
		"roles": roles,
	})
}

// ListAllUsers - List all users with their roles
func (h *RoleAssignmentHandler) ListAllUsers(c *fiber.Ctx) error {
	ctx := context.Background()

	client, err := h.firebaseApp.Auth(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to initialize Firebase Auth",
		})
	}

	// List users (max 1000 at a time)
	iter := client.Users(ctx, "")
	users := []map[string]interface{}{}

	for {
		user, err := iter.Next()
		if err != nil {
			break
		}

		roles := []string{}
		if user.CustomClaims != nil {
			if rolesInterface, ok := user.CustomClaims["roles"]; ok {
				if rolesList, ok := rolesInterface.([]interface{}); ok {
					for _, role := range rolesList {
						if roleStr, ok := role.(string); ok {
							roles = append(roles, roleStr)
						}
					}
				}
			}
		}

		users = append(users, map[string]interface{}{
			"uid":         user.UID,
			"email":       user.Email,
			"displayName": user.DisplayName,
			"roles":       roles,
			"createdAt":   user.UserMetadata.CreationTimestamp,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"users": users,
		"count": len(users),
	})
}

// RemoveRoles - Remove all roles from a user
func (h *RoleAssignmentHandler) RemoveRoles(c *fiber.Ctx) error {
	ctx := context.Background()

	client, err := h.firebaseApp.Auth(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to initialize Firebase Auth",
		})
	}

	email := c.Query("email")
	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email query parameter required",
		})
	}

	user, err := client.GetUserByEmail(ctx, email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Set empty claims
	err = client.SetCustomUserClaims(ctx, user.UID, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove roles",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Roles removed successfully",
		"email":   email,
	})
}

// InitializeSuperAdmin - One-time setup to create the first admin
// Call this endpoint ONCE when you first set up your system
func (h *RoleAssignmentHandler) InitializeSuperAdmin(c *fiber.Ctx) error {
	ctx := context.Background()
	client, err := h.firebaseApp.Auth(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to init Firebase"})
	}

	email := c.Query("email")
	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email is required"})
	}

	user, err := client.GetUserByEmail(ctx, email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	if err := client.SetCustomUserClaims(ctx, user.UID, map[string]interface{}{"roles": []string{"admin"}}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to set admin role"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Super admin initialized",
		"email":   email,
		"roles":   []string{"admin"},
		"note":    "User must re-login to receive new token with roles",
	})
}
