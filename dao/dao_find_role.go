package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"

	"go.mongodb.org/mongo-driver/bson"
)

// DB_FindAdminRole looks up the role of a user in the admin_users collection.
// Searches by email first (new), then falls back to firebaseUid (legacy).
func DB_FindAdminRole(identifier string) (string, bool, error) {
	ctx := context.Background()
	var result struct {
		Role string `bson:"role"`
	}

	// Try email first
	err := dbConfigs.AdminUserCollection.FindOne(ctx, bson.M{"email": identifier}).Decode(&result)
	if err == nil {
		return result.Role, true, nil
	}

	// Fallback: legacy firebaseUid lookup (for existing records)
	err = dbConfigs.AdminUserCollection.FindOne(ctx, bson.M{"firebaseUid": identifier}).Decode(&result)
	if err == nil {
		return result.Role, true, nil
	}

	return "", false, nil // not found is not an error — just unknown
}

// DB_FindMobileUserRole looks up the role of a user across patients and doctor_users.
// Searches by email first, then falls back to firebaseUid for legacy records.
func DB_FindMobileUserRole(identifier string) (string, bool, error) {
	ctx := context.Background()
	var result struct {
		Role string `bson:"role"`
	}

	// Check patients by email
	err := dbConfigs.PatientCollection.FindOne(ctx, bson.M{"email": identifier}).Decode(&result)
	if err == nil {
		return result.Role, true, nil
	}

	// Check patients by legacy firebaseUid
	err = dbConfigs.PatientCollection.FindOne(ctx, bson.M{"firebaseUid": identifier}).Decode(&result)
	if err == nil {
		return result.Role, true, nil
	}

	// Check doctor_users by email
	err = dbConfigs.DoctorUserCollection.FindOne(ctx, bson.M{"email": identifier}).Decode(&result)
	if err == nil {
		return result.Role, true, nil
	}

	// Check doctor_users by legacy firebaseUid
	err = dbConfigs.DoctorUserCollection.FindOne(ctx, bson.M{"firebaseUid": identifier}).Decode(&result)
	if err == nil {
		return result.Role, true, nil
	}

	return "", false, nil // not found in either collection
}
