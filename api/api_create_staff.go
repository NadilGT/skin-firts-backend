package api

import (
	"context"
	"fmt"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/utils"
	"log"
	"time"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
)

type CreateStaffRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	Role        string `json:"role"` // "admin", "doctor", "cashier", "receptionist", etc.
}

type StaffHandler struct {
	firebaseApp *firebase.App
}

func NewStaffHandler(app *firebase.App) *StaffHandler {
	return &StaffHandler{firebaseApp: app}
}

// POST /admin/create-staff
// Super Admin only: Creates a Firebase Auth user, assigns custom claims (role),
// sends a password reset email to set their own password, and saves the user
// data into the corresponding database collection.
func (h *StaffHandler) CreateStaffAccount(c *fiber.Ctx) error {
	var req CreateStaffRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" || req.Name == "" || req.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email, name, and role are required",
		})
	}

	ctx := context.Background()
	client, err := h.firebaseApp.Auth(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect to Firebase Auth"})
	}

	// 1. Create the user in Firebase Auth with a default random password
	defaultPassword := "TeMp@123!" // The user will be forced to change it via the password reset email
	params := (&auth.UserToCreate{}).
		Email(req.Email).
		Password(defaultPassword).
		DisplayName(req.Name)

	if req.PhoneNumber != "" {
		params = params.PhoneNumber(req.PhoneNumber)
	}

	fwUser, err := client.CreateUser(ctx, params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to create Firebase Auth user: %v", err)})
	}

	// 2. Assign Custom Claims (Roles)
	claims := map[string]interface{}{
		"roles": []string{req.Role},
	}
	if err := client.SetCustomUserClaims(ctx, fwUser.UID, claims); err != nil {
		// Rollback Auth user creation if we can't set claims
		_ = client.DeleteUser(ctx, fwUser.UID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to assign roles. User creation rolled back."})
	}

	// 3. Generate Password Reset Link and Send Email
	resetLink, err := client.PasswordResetLink(ctx, req.Email)
	if err != nil {
		log.Println("⚠️ Failed to generate password reset link:", err)
		// We won't rollback here, but we'll notify that email sending failed
	} else {
		// Send the email asynchronously so it doesn't block the API response
		go func(email, name, role, link string) {
			emailSubject := "Welcome to the Med Center Portal - Set up your account"
			emailBody := fmt.Sprintf(`
				<h2>Hello %s!</h2>
				<p>An account has been created for you as a <strong>%s</strong>.</p>
				<p>Please click the link below to set your password and access the portal:</p>
				<a href="%s" style="padding:10px 15px;background-color:#007bff;color:white;text-decoration:none;border-radius:5px;">Set Password</a>
				<br><br>
				<p>If the button doesn't work, copy and paste this link into your browser:</p>
				<p>%s</p>
				<br>
				<p>Thanks,<br>Management Team</p>
			`, name, role, link, link)

			if err := utils.SendEmail([]string{email}, emailSubject, emailBody); err != nil {
				log.Println("⚠️ Failed to send password reset email:", err)
			} else {
				log.Println("✅ Password reset email sent securely to:", email)
			}
		}(req.Email, req.Name, req.Role, resetLink)
	}

	// 4. Save to MongoDB
	var dbUserID string
	var dbErr error

	switch req.Role {
	case "admin", "super_admin":
		dbUserID, dbErr = dao.GenerateId(ctx, "admin_users", "AD")
		if dbErr == nil {
			admin := dto.AdminUser{
				UserID:      dbUserID,
				FirebaseUID: fwUser.UID,
				Name:        req.Name,
				Email:       req.Email,
				PhoneNumber: req.PhoneNumber,
				Role:        req.Role,
				CreatedAt:   time.Now(),
			}
			dbErr = dao.DB_CreateAdminUser(admin)
		}

	case "doctor":
		dbUserID, dbErr = dao.GenerateId(ctx, "doctor_users", "DOC")
		if dbErr == nil {
			doctor := dto.DoctorUser{
				UserID:      dbUserID,
				FirebaseUID: fwUser.UID,
				Name:        req.Name,
				Email:       req.Email,
				PhoneNumber: req.PhoneNumber,
				Role:        req.Role,
				CreatedAt:   time.Now(),
			}
			dbErr = dao.DB_CreateDoctorUser(doctor)
		}

	default: // cashiers, receptionists, other staff
		dbUserID, dbErr = dao.GenerateId(ctx, "staff_users", "STF")
		if dbErr == nil {
			staff := dto.StaffUser{
				UserID:      dbUserID,
				FirebaseUID: fwUser.UID,
				Name:        req.Name,
				Email:       req.Email,
				PhoneNumber: req.PhoneNumber,
				Role:        req.Role,
				CreatedAt:   time.Now(),
			}
			dbErr = dao.DB_CreateStaffUser(staff)
		}
	}

	if dbErr != nil {
		// Rollback Auth User if DB creation fails
		_ = client.DeleteUser(ctx, fwUser.UID)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to save user in database: %v. Account creation rolled back.", dbErr),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Account successfully created. An email has been sent for password setup.",
		"userId":    dbUserID,
		"uid":       fwUser.UID,
		"email":     req.Email,
		"role":      req.Role,
		"resetLink": resetLink, // Note: returning this just for easy local debugging if needed. Can be omitted in prod.
	})
}

// GET /admin/search-staff
// Super Admin only: Searches for staff members across all categories with pagination.
func (h *StaffHandler) SearchStaff(c *fiber.Ctx) error {
	var query dto.SearchStaffQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid query parameters: " + err.Error(),
		})
	}

	// Default pagination
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
