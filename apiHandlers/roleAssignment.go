// apiHandlers/roleAssignment.go
package apiHandlers

import (
	"context"
	"log"
	"os"

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
// Call this function when the server starts
func InitializeSuperAdmin(app *firebase.App) {
	ctx := context.Background()

	email := os.Getenv("SUPER_ADMIN_EMAIL")
	if email == "" {
		log.Println("No super admin email configured")
		return
	}

	client, err := app.Auth(ctx)
	if err != nil {
		log.Println("❌ Failed to initialize Firebase Auth for Super Admin check:", err)
		return
	}

	user, err := client.GetUserByEmail(ctx, email)
	if err != nil {
		log.Printf("⚠️ User not found for super admin (%s). Please sign up first.\n", email)
		return
	}

	// Check if already has roles
	if user.CustomClaims != nil {
		if _, ok := user.CustomClaims["roles"]; ok {
			log.Println("✅ Super admin already initialized:", email)
			return
		}
	}

	// Assign admin role
	if err := client.SetCustomUserClaims(ctx, user.UID, map[string]interface{}{"roles": []string{"super_admin"}}); err != nil {
		log.Println("❌ Failed to set super_admin role:", err)
		return
	}

	log.Println("🚀 Super admin initialized successfully:", email)
}
