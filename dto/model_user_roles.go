package dto

import "time"

// UserRole constants
const (
	RolePatient     = "patient"
	RoleDoctor      = "doctor"
	RoleAdmin       = "admin"
	RoleSuperAdmin  = "super_admin"
	RolePharmacist  = "pharmacist"
	RoleReceptionist = "receptionist"
)

// PatientUser represents a registered patient stored in the "patients" collection.
type PatientUser struct {
	UserID      string    `json:"userId"      bson:"userId"`
	FirebaseUID string    `json:"firebaseUid" bson:"firebaseUid"`
	Name        string    `json:"name"        bson:"name"`
	Email       string    `json:"email"       bson:"email"`
	PhoneNumber string    `json:"phoneNumber" bson:"phoneNumber"`
	Role        string    `json:"role"        bson:"role"`
	FcmToken    string    `json:"fcmToken,omitempty" bson:"fcmToken,omitempty"`
	CreatedAt   time.Time `json:"createdAt"   bson:"createdAt"`
}

// DoctorUser represents a registered doctor stored in the "doctor_users" collection.
// Note: this is separate from DoctorInfoModel which holds clinical profile data.
type DoctorUser struct {
	UserID      string    `json:"userId"      bson:"userId"`
	FirebaseUID string    `json:"firebaseUid" bson:"firebaseUid"`
	Name        string    `json:"name"        bson:"name"`
	Email       string    `json:"email"       bson:"email"`
	PhoneNumber string    `json:"phoneNumber" bson:"phoneNumber"`
	Role        string    `json:"role"        bson:"role"`
	BranchId    string    `json:"branchId,omitempty" bson:"branchId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"   bson:"createdAt"`
}

// AdminUser represents a registered admin stored in the "admin_users" collection.
type AdminUser struct {
	UserID      string    `json:"userId"      bson:"userId"`
	FirebaseUID string    `json:"firebaseUid" bson:"firebaseUid"`
	Name        string    `json:"name"        bson:"name"`
	Email       string    `json:"email"       bson:"email"`
	PhoneNumber string    `json:"phoneNumber" bson:"phoneNumber"`
	Role        string    `json:"role"        bson:"role"`
	BranchId    string    `json:"branchId,omitempty" bson:"branchId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"   bson:"createdAt"`
}

// StaffUser represents other staff members (cashiers, receptionists, etc.) stored in the "staff_users" collection.
type StaffUser struct {
	UserID      string    `json:"userId"      bson:"userId"`
	FirebaseUID string    `json:"firebaseUid" bson:"firebaseUid"`
	Name        string    `json:"name"        bson:"name"`
	Email       string    `json:"email"       bson:"email"`
	PhoneNumber string    `json:"phoneNumber" bson:"phoneNumber"`
	Role        string    `json:"role"        bson:"role"`
	BranchId    string    `json:"branchId,omitempty" bson:"branchId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"   bson:"createdAt"`
}

// RegisterUserRequest is the common request body for all 3 registration endpoints.
// The caller must pass the Firebase UID obtained after mobile sign-up.
type RegisterUserRequest struct {
	FirebaseUID string `json:"firebaseUid"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	BranchId    string `json:"branchId,omitempty"`   // optional — required for staff/admin
}

// SearchStaffQuery represents the query parameters for searching staff.
type SearchStaffQuery struct {
	Query string `json:"query" query:"query"`
	Role  string `json:"role" query:"role"`
	Page  int    `json:"page" query:"page"`
	Limit int    `json:"limit" query:"limit"`
}

// StaffMember is a unified representation of any staff/admin/doctor user for search results.
type StaffMember struct {
	UserID      string    `json:"userId"      bson:"userId"`
	FirebaseUID string    `json:"firebaseUid" bson:"firebaseUid"`
	Name        string    `json:"name"        bson:"name"`
	Email       string    `json:"email"       bson:"email"`
	PhoneNumber string    `json:"phoneNumber" bson:"phoneNumber"`
	Role        string    `json:"role"        bson:"role"`
	BranchId    string    `json:"branchId,omitempty" bson:"branchId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"   bson:"createdAt"`
}

// StaffSearchResponse represents the paginated response for a staff search.
type StaffSearchResponse struct {
	Data       []StaffMember `json:"data"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"totalPages"`
}

// SearchPatientQuery represents the query parameters for searching patients.
type SearchPatientQuery struct {
	Query string `json:"query" query:"query"`
	Page  int    `json:"page" query:"page"`
	Limit int    `json:"limit" query:"limit"`
}

// PatientSearchResponse represents the paginated response for a patient search.
type PatientSearchResponse struct {
	Data       []PatientUser `json:"data"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"totalPages"`
}
