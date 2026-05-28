package apiHandlers

import (
	"lawyerSL-Backend/auth"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware wraps the local JWT middleware, replacing the old Firebase
// token verification.  All Fiber Locals set by auth.JWTMiddleware() remain
// identical to the old Firebase middleware so no downstream handler changes
// are needed:
//
//	c.Locals("userId", ...)    — app-level user ID
//	c.Locals("uid", ...)       — backward-compat alias for userId
//	c.Locals("email", ...)
//	c.Locals("role", ...)      — primary role
//	c.Locals("roles", ...)     — []string of all roles
//	c.Locals("branchId", ...)
type AuthMiddleware struct{}

// NewAuthMiddleware constructs an AuthMiddleware. The authConfig and any
// legacy firebase parameters are ignored — kept only for call-site compatibility
// during transition.
func NewAuthMiddleware(args ...interface{}) *AuthMiddleware {
	return &AuthMiddleware{}
}

// ValidateToken is the Fiber middleware handler. Delegates to auth.JWTMiddleware.
func (a *AuthMiddleware) ValidateToken(c *fiber.Ctx) error {
	return auth.JWTMiddleware()(c)
}

// RequiresRole is kept here as a convenience re-export so existing router.go
// call sites (RequiresRole("admin")) continue to compile unchanged.
func RequiresRole(requiredRole string) fiber.Handler {
	return auth.RequiresRole(requiredRole)
}
