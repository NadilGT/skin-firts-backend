package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// JWTMiddleware validates the Bearer token, extracts claims, and injects
// them into Fiber's context locals for downstream handlers.
//
// Locals set:
//   - "userId"   — the app-level user ID (e.g. AD-001)
//   - "uid"      — alias for userId (backward-compat with legacy handlers)
//   - "email"    — user's email
//   - "role"     — primary role string
//   - "roles"    — []string of all roles
//   - "branchId" — assigned branch
func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing Authorization header",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid Authorization header format. Expected: Bearer <token>",
			})
		}

		claims, err := ParseJWT(parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Primary locals
		c.Locals("userId", claims.UserId)
		c.Locals("uid", claims.UserId) // backward-compat alias
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		c.Locals("roles", claims.Roles)
		c.Locals("branchId", claims.BranchId)

		return c.Next()
	}
}

// RequiresRole returns a middleware that blocks requests whose token does not
// carry the required role. super_admin bypasses all role checks.
func RequiresRole(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rolesInterface := c.Locals("roles")
		if rolesInterface == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Roles not found in context — token missing?",
			})
		}
		roles, ok := rolesInterface.([]string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid roles format in token",
			})
		}

		for _, r := range roles {
			if r == "super_admin" || r == requiredRole {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied. Required role: " + requiredRole,
		})
	}
}
