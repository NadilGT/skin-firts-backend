package api

import (
	"context"
	"lawyerSL-Backend/auth"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/gofiber/fiber/v2"
)

// CreateStaffRequest is the body for POST /admin/create-staff.
type CreateStaffRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`    // admin provides the initial password
	PhoneNumber string `json:"phoneNumber"`
	Role        string `json:"role"`     // "admin" | "doctor" | "staff" | etc.
	BranchIds   []string `json:"branchIds"`
}

// StaffHandler handles staff creation and search.
type StaffHandler struct{}

func NewStaffHandler(_ ...interface{}) *StaffHandler {
	return &StaffHandler{}
}

// CreateStaffAccount handles POST /admin/create-staff.
// Admin provides email + password + role.  Password is bcrypt-hashed and
// stored in MongoDB.  No Firebase involvement.
//
// The MustChangePassword flag is set to true so the frontend can force the
// staff member to change their password on first login.
func (h *StaffHandler) CreateStaffAccount(c *fiber.Ctx) error {
	var req CreateStaffRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" || req.Name == "" || req.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email, name, and role are required",
		})
	}

	if req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password is required (admin provides the initial password for the account)",
		})
	}

	// For CreateStaffRequest we now support multiple branches, but ResolveBranchId might need a single ID.
	// For now we trust the BranchIds array passed or set a default if empty.
	var finalBranchIds []string
	if len(req.BranchIds) > 0 {
		finalBranchIds = req.BranchIds
	} else {
		// Fallback to resolve a default branch
		branchId, err := ResolveBranchId(c, "")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		finalBranchIds = []string{branchId}
	}

	// Hash the provided password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	ctx := context.Background()
	now := time.Now()
	var dbUserID string
	var dbErr error

	switch req.Role {
	case "admin", "super_admin":
		dbUserID, dbErr = dao.GenerateId(ctx, "admin_users", "AD")
		if dbErr == nil {
			admin := dto.AdminUser{
				UserID:             dbUserID,
				Name:               req.Name,
				Email:              req.Email,
				PasswordHash:       passwordHash,
				PhoneNumber:        req.PhoneNumber,
				Role:               req.Role,
				BranchIds:          finalBranchIds,
				Status:             dto.StatusActive,
				MustChangePassword: true, // force change on first login
				CreatedAt:          now,
			}
			dbErr = dao.DB_CreateAdminUser(admin)
		}

	case "doctor":
		dbUserID, dbErr = dao.GenerateId(ctx, "doctor_users", "DOC")
		if dbErr == nil {
			doctor := dto.DoctorUser{
				UserID:             dbUserID,
				Name:               req.Name,
				Email:              req.Email,
				PasswordHash:       passwordHash,
				PhoneNumber:        req.PhoneNumber,
				Role:               req.Role,
				BranchIds:          finalBranchIds,
				Status:             dto.StatusActive,
				MustChangePassword: true,
				CreatedAt:          now,
			}
			dbErr = dao.DB_CreateDoctorUser(doctor)
		}

	default: // cashier, receptionist, pharmacist, staff, etc.
		dbUserID, dbErr = dao.GenerateId(ctx, "staff_users", "STF")
		if dbErr == nil {
			staff := dto.StaffUser{
				UserID:             dbUserID,
				Name:               req.Name,
				Email:              req.Email,
				PasswordHash:       passwordHash,
				PhoneNumber:        req.PhoneNumber,
				Role:               req.Role,
				BranchIds:          finalBranchIds,
				Status:             dto.StatusActive,
				MustChangePassword: true,
				CreatedAt:          now,
			}
			dbErr = dao.DB_CreateStaffUser(staff)
		}
	}

	if dbErr != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": dbErr.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":            "Staff account created successfully",
		"userId":             dbUserID,
		"email":              req.Email,
		"role":               req.Role,
		"branchIds":          finalBranchIds,
		"mustChangePassword": true,
		"note":               "Staff must change their password on first login",
	})
}

// SearchStaff handles GET /admin/search-staff
func (h *StaffHandler) SearchStaff(c *fiber.Ctx) error {
	var query dto.SearchStaffQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid query parameters: " + err.Error(),
		})
	}

	branchId, err := ResolveBranchId(c, query.BranchId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	query.BranchId = branchId

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	staff, total, err := dao.DB_SearchStaff(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search staff: " + err.Error(),
		})
	}

	totalPages := (total + int64(query.Limit) - 1) / int64(query.Limit)

	return c.Status(fiber.StatusOK).JSON(dto.StaffSearchResponse{
		Data:       staff,
		Total:      total,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: int(totalPages),
	})
}
