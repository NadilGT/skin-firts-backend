package dto

import "time"

// ---------------------------------------------------------------------------
// Role constants — both lowercase (DB-stored) and uppercase (API-friendly)
// ---------------------------------------------------------------------------

const (
	RolePatient      = "patient"
	RoleDoctor       = "doctor"
	RoleAdmin        = "admin"
	RoleSuperAdmin   = "super_admin"
	RolePharmacist   = "pharmacist"
	RoleReceptionist = "receptionist"
	RoleStaff        = "staff"
)

// ---------------------------------------------------------------------------
// Account status constants
// ---------------------------------------------------------------------------

const (
	StatusActive    = "ACTIVE"
	StatusInactive  = "INACTIVE"
	StatusSuspended = "SUSPENDED"
)

// ---------------------------------------------------------------------------
// User Models
// ---------------------------------------------------------------------------

// PatientUser represents a registered patient stored in the "patients" collection.
type PatientUser struct {
	UserID             string    `json:"userId"      bson:"userId"`
	FirebaseUID        string    `json:"firebaseUid,omitempty" bson:"firebaseUid,omitempty"` // kept for backward-compat
	Name               string    `json:"name"        bson:"name"`
	Email              string    `json:"email"       bson:"email"`
	PasswordHash       string    `json:"-"           bson:"passwordHash"` // NEVER serialised to JSON
	PhoneNumber        string    `json:"phoneNumber" bson:"phoneNumber"`
	Role               string    `json:"role"        bson:"role"`
	Status             string    `json:"status"      bson:"status"`
	MustChangePassword bool      `json:"mustChangePassword" bson:"mustChangePassword"`
	FcmToken           string    `json:"fcmToken,omitempty" bson:"fcmToken,omitempty"`
	CreatedAt          time.Time `json:"createdAt"   bson:"createdAt"`
}

// DoctorUser represents a registered doctor stored in the "doctor_users" collection.
// Note: separate from DoctorInfoModel which holds clinical profile data.
type DoctorUser struct {
	UserID             string    `json:"userId"      bson:"userId"`
	FirebaseUID        string    `json:"firebaseUid,omitempty" bson:"firebaseUid,omitempty"` // kept for backward-compat
	Name               string    `json:"name"        bson:"name"`
	Email              string    `json:"email"       bson:"email"`
	PasswordHash       string    `json:"-"           bson:"passwordHash"` // NEVER serialised to JSON
	PhoneNumber        string    `json:"phoneNumber" bson:"phoneNumber"`
	Role               string    `json:"role"        bson:"role"`
	BranchId           string    `json:"branchId,omitempty" bson:"branchId,omitempty"`
	Status             string    `json:"status"      bson:"status"`
	MustChangePassword bool      `json:"mustChangePassword" bson:"mustChangePassword"`
	CreatedAt          time.Time `json:"createdAt"   bson:"createdAt"`
}

// AdminUser represents a registered admin stored in the "admin_users" collection.
type AdminUser struct {
	UserID             string    `json:"userId"      bson:"userId"`
	FirebaseUID        string    `json:"firebaseUid,omitempty" bson:"firebaseUid,omitempty"` // kept for backward-compat
	Name               string    `json:"name"        bson:"name"`
	Email              string    `json:"email"       bson:"email"`
	PasswordHash       string    `json:"-"           bson:"passwordHash"` // NEVER serialised to JSON
	PhoneNumber        string    `json:"phoneNumber" bson:"phoneNumber"`
	Role               string    `json:"role"        bson:"role"`
	BranchId           string    `json:"branchId,omitempty" bson:"branchId,omitempty"`
	Status             string    `json:"status"      bson:"status"`
	MustChangePassword bool      `json:"mustChangePassword" bson:"mustChangePassword"`
	CreatedAt          time.Time `json:"createdAt"   bson:"createdAt"`
}

// StaffUser represents other staff members stored in the "staff_users" collection.
type StaffUser struct {
	UserID             string    `json:"userId"      bson:"userId"`
	FirebaseUID        string    `json:"firebaseUid,omitempty" bson:"firebaseUid,omitempty"` // kept for backward-compat
	Name               string    `json:"name"        bson:"name"`
	Email              string    `json:"email"       bson:"email"`
	PasswordHash       string    `json:"-"           bson:"passwordHash"` // NEVER serialised to JSON
	PhoneNumber        string    `json:"phoneNumber" bson:"phoneNumber"`
	Role               string    `json:"role"        bson:"role"`
	BranchId           string    `json:"branchId,omitempty" bson:"branchId,omitempty"`
	Status             string    `json:"status"      bson:"status"`
	MustChangePassword bool      `json:"mustChangePassword" bson:"mustChangePassword"`
	CreatedAt          time.Time `json:"createdAt"   bson:"createdAt"`
}

// ---------------------------------------------------------------------------
// Shared Request/Response Types
// ---------------------------------------------------------------------------

// RegisterUserRequest is the common request body for the legacy registration
// endpoints (/register/patient, /register/doctor-user, /register/admin).
// FirebaseUID kept as omitempty for backward-compat with existing frontends.
type RegisterUserRequest struct {
	FirebaseUID string `json:"firebaseUid,omitempty"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phoneNumber"`
	BranchId    string `json:"branchId,omitempty"`
}

// SearchStaffQuery represents the query parameters for searching staff.
type SearchStaffQuery struct {
	Query    string `json:"query"    query:"query"`
	Role     string `json:"role"     query:"role"`
	BranchId string `json:"branchId" query:"branchId"`
	Page     int    `json:"page"     query:"page"`
	Limit    int    `json:"limit"    query:"limit"`
}

// StaffMember is a unified representation of any staff/admin/doctor user for search results.
// PasswordHash is deliberately omitted.
type StaffMember struct {
	UserID      string    `json:"userId"      bson:"userId"`
	FirebaseUID string    `json:"firebaseUid,omitempty" bson:"firebaseUid,omitempty"`
	Name        string    `json:"name"        bson:"name"`
	Email       string    `json:"email"       bson:"email"`
	PhoneNumber string    `json:"phoneNumber" bson:"phoneNumber"`
	Role        string    `json:"role"        bson:"role"`
	BranchId    string    `json:"branchId,omitempty" bson:"branchId,omitempty"`
	Status      string    `json:"status"      bson:"status"`
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
	Page  int    `json:"page"  query:"page"`
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
