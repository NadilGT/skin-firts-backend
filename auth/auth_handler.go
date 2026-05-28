package auth

import (
	"lawyerSL-Backend/dao"

	"github.com/gofiber/fiber/v2"
)

// ---------------------------------------------------------------------------
// POST /auth/login
// ---------------------------------------------------------------------------

// Login accepts email + password, verifies against MongoDB, and returns a JWT.
//
// Response: { token, user: { id, name, email, role, branchId, status, mustChangePassword } }
func Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	user, err := FindUserByEmail(req.Email)
	if err != nil {
		// Generic message — never reveal whether the email exists
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Check account status before attempting password comparison
	if user.Status == "INACTIVE" || user.Status == "SUSPENDED" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Account is " + user.Status + ". Contact your administrator.",
		})
	}

	// Legacy users created before the password migration may have no hash
	if user.PasswordHash == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "No password set for this account. Ask an admin to set one via /auth/set-password",
		})
	}

	if !CheckPassword(user.PasswordHash, req.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	token, err := GenerateJWT(user.UserId, user.Role, user.BranchId, user.Email, user.Roles)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.Status(fiber.StatusOK).JSON(LoginResponse{
		Token: token,
		User: AuthUserResponse{
			ID:                 user.UserId,
			Name:               user.Name,
			Email:              user.Email,
			Role:               user.Role,
			BranchId:           user.BranchId,
			Status:             user.Status,
			MustChangePassword: user.MustChangePassword,
		},
	})
}

// ---------------------------------------------------------------------------
// POST /auth/register
// ---------------------------------------------------------------------------

// Register creates a new user account with a hashed password.
// Intended for self-registration (patients) or admin-driven onboarding.
func Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user, err := RegisterUser(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	token, err := GenerateJWT(user.UserId, user.Role, user.BranchId, user.Email, user.Roles)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "User created but failed to generate token",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(LoginResponse{
		Token: token,
		User: AuthUserResponse{
			ID:                 user.UserId,
			Name:               user.Name,
			Email:              user.Email,
			Role:               user.Role,
			BranchId:           user.BranchId,
			Status:             user.Status,
			MustChangePassword: user.MustChangePassword,
		},
	})
}

// ---------------------------------------------------------------------------
// GET /auth/me
// ---------------------------------------------------------------------------

// Me returns the authenticated user's identity extracted from their JWT claims.
// Also reads the live DB record so status/mustChangePassword are up-to-date.
func Me(c *fiber.Ctx) error {
	userId, _ := c.Locals("userId").(string)
	email, _ := c.Locals("email").(string)
	branchId, _ := c.Locals("branchId").(string)
	role, _ := c.Locals("role").(string)
	roles, _ := c.Locals("roles").([]string)

	isSuperAdmin := false
	for _, r := range roles {
		if r == "super_admin" {
			isSuperAdmin = true
			break
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"userId":       userId,
		"uid":          userId, // backward-compat alias
		"email":        email,
		"branchId":     branchId,
		"role":         role,
		"roles":        roles,
		"isSuperAdmin": isSuperAdmin,
	})
}

// ---------------------------------------------------------------------------
// POST /auth/set-password  (admin-only)
// ---------------------------------------------------------------------------

// SetPassword allows an admin to set or reset any user's password.
// This is the migration helper for existing users who have no passwordHash.
func SetPassword(c *fiber.Ctx) error {
	var req SetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email and newPassword are required",
		})
	}

	user, err := FindUserByEmail(req.Email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found with that email",
		})
	}

	hash, err := HashPassword(req.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	if err := dao.DB_SetPasswordHash(req.Email, user.Collection, hash); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update password: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password updated successfully for " + req.Email,
	})
}

// ---------------------------------------------------------------------------
// POST /auth/change-password  (authenticated user)
// ---------------------------------------------------------------------------

// ChangePassword lets a logged-in user update their own password by supplying
// the current password first. Also clears the mustChangePassword flag.
func ChangePassword(c *fiber.Ctx) error {
	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "oldPassword and newPassword are required",
		})
	}

	email, _ := c.Locals("email").(string)
	if email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Cannot determine current user",
		})
	}

	user, err := FindUserByEmail(email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	if !CheckPassword(user.PasswordHash, req.OldPassword) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Current password is incorrect",
		})
	}

	if req.OldPassword == req.NewPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "New password must be different from the current password",
		})
	}

	hash, err := HashPassword(req.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash new password",
		})
	}

	if err := dao.DB_SetPasswordHash(email, user.Collection, hash); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save new password",
		})
	}

	// Clear the mustChangePassword flag
	_ = dao.DB_ClearMustChangePassword(email, user.Collection)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password changed successfully",
	})
}
