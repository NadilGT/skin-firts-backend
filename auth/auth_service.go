package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ---------------------------------------------------------------------------
// Unified user record returned by FindUserByEmail
// ---------------------------------------------------------------------------

// FoundUser is an internal structure used across auth service functions.
// It gathers the common fields from any user collection.
type FoundUser struct {
	UserId             string
	Name               string
	Email              string
	PasswordHash       string
	Role               string
	Roles              []string
	BranchId           string
	Status             string
	MustChangePassword bool
	Collection         string // which collection the user lives in
}

// ---------------------------------------------------------------------------
// FindUserByEmail — searches all 4 user collections
// ---------------------------------------------------------------------------

// FindUserByEmail searches admin_users → doctor_users → staff_users → patients
// and returns the first match as a FoundUser.
func FindUserByEmail(email string) (*FoundUser, error) {
	ctx := context.Background()

	type rawUser struct {
		UserID             string `bson:"userId"`
		Name               string `bson:"name"`
		Email              string `bson:"email"`
		PasswordHash       string `bson:"passwordHash"`
		Role               string `bson:"role"`
		BranchId           string `bson:"branchId"`
		Status             string `bson:"status"`
		MustChangePassword bool   `bson:"mustChangePassword"`
	}

	collections := []struct {
		col  *mongo.Collection
		name string
	}{
		{dbConfigs.AdminUserCollection, "admin_users"},
		{dbConfigs.DoctorUserCollection, "doctor_users"},
		{dbConfigs.StaffUserCollection, "staff_users"},
		{dbConfigs.PatientCollection, "patients"},
	}

	for _, entry := range collections {
		var raw rawUser
		err := entry.col.FindOne(ctx, bson.M{"email": email}).Decode(&raw)
		if err == nil {
			roles := []string{raw.Role}
			status := raw.Status
			if status == "" {
				status = "ACTIVE" // default for legacy records
			}
			return &FoundUser{
				UserId:             raw.UserID,
				Name:               raw.Name,
				Email:              raw.Email,
				PasswordHash:       raw.PasswordHash,
				Role:               raw.Role,
				Roles:              roles,
				BranchId:           raw.BranchId,
				Status:             status,
				MustChangePassword: raw.MustChangePassword,
				Collection:         entry.name,
			}, nil
		}
	}

	return nil, errors.New("user not found")
}

// ---------------------------------------------------------------------------
// RegisterUser — hash password and insert into correct collection
// ---------------------------------------------------------------------------

// RegisterUser hashes the plaintext password and writes the user record into
// the appropriate MongoDB collection based on role.
func RegisterUser(req RegisterRequest) (*FoundUser, error) {
	ctx := context.Background()

	if req.Email == "" || req.Password == "" || req.Name == "" {
		return nil, errors.New("name, email, and password are required")
	}
	if req.Role == "" {
		return nil, errors.New("role is required")
	}

	// Reject if email already exists anywhere
	existing, _ := FindUserByEmail(req.Email)
	if existing != nil {
		return nil, fmt.Errorf("a user with email %s already exists", req.Email)
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()

	switch req.Role {
	case "super_admin", "admin":
		userID, err := dao.GenerateId(ctx, "admin_users", "AD")
		if err != nil {
			return nil, err
		}
		admin := dto.AdminUser{
			UserID:             userID,
			Name:               req.Name,
			Email:              req.Email,
			PasswordHash:       hash,
			PhoneNumber:        req.PhoneNumber,
			Role:               req.Role,
			BranchId:           req.BranchId,
			Status:             "ACTIVE",
			MustChangePassword: false,
			CreatedAt:          now,
		}
		if err := dao.DB_CreateAdminUser(admin); err != nil {
			return nil, err
		}
		return &FoundUser{
			UserId: userID, Name: req.Name, Email: req.Email,
			Role: req.Role, Roles: []string{req.Role}, BranchId: req.BranchId,
			Status: "ACTIVE", MustChangePassword: false, Collection: "admin_users",
		}, nil

	case "doctor":
		userID, err := dao.GenerateId(ctx, "doctor_users", "DOC")
		if err != nil {
			return nil, err
		}
		doctor := dto.DoctorUser{
			UserID:             userID,
			Name:               req.Name,
			Email:              req.Email,
			PasswordHash:       hash,
			PhoneNumber:        req.PhoneNumber,
			Role:               req.Role,
			BranchId:           req.BranchId,
			Status:             "ACTIVE",
			MustChangePassword: false,
			CreatedAt:          now,
		}
		if err := dao.DB_CreateDoctorUser(doctor); err != nil {
			return nil, err
		}
		return &FoundUser{
			UserId: userID, Name: req.Name, Email: req.Email,
			Role: req.Role, Roles: []string{req.Role}, BranchId: req.BranchId,
			Status: "ACTIVE", MustChangePassword: false, Collection: "doctor_users",
		}, nil

	case "patient":
		userID, err := dao.GenerateId(ctx, "patients", "PAT")
		if err != nil {
			return nil, err
		}
		patient := dto.PatientUser{
			UserID:             userID,
			Name:               req.Name,
			Email:              req.Email,
			PasswordHash:       hash,
			PhoneNumber:        req.PhoneNumber,
			Role:               req.Role,
			Status:             "ACTIVE",
			MustChangePassword: false,
			CreatedAt:          now,
		}
		if err := dao.DB_CreatePatient(patient); err != nil {
			return nil, err
		}
		return &FoundUser{
			UserId: userID, Name: req.Name, Email: req.Email,
			Role: req.Role, Roles: []string{req.Role},
			Status: "ACTIVE", MustChangePassword: false, Collection: "patients",
		}, nil

	default:
		// All other staff roles (receptionist, pharmacist, cashier, etc.)
		userID, err := dao.GenerateId(ctx, "staff_users", "STF")
		if err != nil {
			return nil, err
		}
		staff := dto.StaffUser{
			UserID:             userID,
			Name:               req.Name,
			Email:              req.Email,
			PasswordHash:       hash,
			PhoneNumber:        req.PhoneNumber,
			Role:               req.Role,
			BranchId:           req.BranchId,
			Status:             "ACTIVE",
			MustChangePassword: false,
			CreatedAt:          now,
		}
		if err := dao.DB_CreateStaffUser(staff); err != nil {
			return nil, err
		}
		return &FoundUser{
			UserId: userID, Name: req.Name, Email: req.Email,
			Role: req.Role, Roles: []string{req.Role}, BranchId: req.BranchId,
			Status: "ACTIVE", MustChangePassword: false, Collection: "staff_users",
		}, nil
	}
}

// ---------------------------------------------------------------------------
// InitializeSuperAdmin — run on startup to seed first super admin
// ---------------------------------------------------------------------------

// InitializeSuperAdmin checks for an existing super admin in admin_users.
// If none is found, it creates one using SUPER_ADMIN_EMAIL + SUPER_ADMIN_PASSWORD.
func InitializeSuperAdmin() {
	email := os.Getenv("SUPER_ADMIN_EMAIL")
	password := os.Getenv("SUPER_ADMIN_PASSWORD")

	if email == "" || password == "" {
		log.Println("⚠️  SUPER_ADMIN_EMAIL or SUPER_ADMIN_PASSWORD not set — skipping super admin seed")
		return
	}

	existing, _ := FindUserByEmail(email)
	if existing != nil {
		log.Printf("✅ Super admin already exists: %s (role=%s)\n", email, existing.Role)
		return
	}

	log.Printf("🚀 Creating super admin: %s\n", email)

	hash, err := HashPassword(password)
	if err != nil {
		log.Println("❌ Failed to hash super admin password:", err)
		return
	}

	ctx := context.Background()
	userID, err := dao.GenerateId(ctx, "admin_users", "AD")
	if err != nil {
		log.Println("❌ Failed to generate super admin ID:", err)
		return
	}

	admin := dto.AdminUser{
		UserID:             userID,
		Name:               "Super Admin",
		Email:              email,
		PasswordHash:       hash,
		Role:               "super_admin",
		BranchId:           "BRN-001",
		Status:             "ACTIVE",
		MustChangePassword: false,
		CreatedAt:          time.Now(),
	}

	if err := dao.DB_CreateAdminUser(admin); err != nil {
		log.Println("❌ Failed to create super admin:", err)
		return
	}

	log.Printf("✅ Super admin created successfully: %s (ID=%s)\n", email, userID)
}
