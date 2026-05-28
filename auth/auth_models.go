package auth

import "github.com/golang-jwt/jwt/v5"

// ---------------------------------------------------------------------------
// JWT Claims
// ---------------------------------------------------------------------------

// JWTClaims holds all data embedded inside each access token.
type JWTClaims struct {
	UserId   string   `json:"userId"`
	Role     string   `json:"role"`
	Roles    []string `json:"roles"`
	BranchIds []string `json:"branchIds"`
	Email    string   `json:"email"`
	jwt.RegisteredClaims
}

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// LoginRequest is the body accepted by POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest is the body accepted by POST /auth/register.
// Role must be one of: super_admin | admin | doctor | staff | patient
type RegisterRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phoneNumber"`
	Role        string `json:"role"`
	BranchIds   []string `json:"branchIds"`
}

// ChangePasswordRequest is the body for POST /auth/change-password.
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// SetPasswordRequest is the body for POST /auth/set-password (admin only).
type SetPasswordRequest struct {
	Email       string `json:"email"`
	NewPassword string `json:"newPassword"`
}

// ---------------------------------------------------------------------------
// Response DTOs
// ---------------------------------------------------------------------------

// AuthUserResponse is the sanitised user payload returned in auth responses.
// NEVER include PasswordHash here.
type AuthUserResponse struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Email             string `json:"email"`
	Role              string `json:"role"`
	BranchIds         []string `json:"branchIds,omitempty"`
	Status            string `json:"status"`
	MustChangePassword bool  `json:"mustChangePassword"`
}

// LoginResponse is the full response body for a successful login.
type LoginResponse struct {
	Token string           `json:"token"`
	User  AuthUserResponse `json:"user"`
}
