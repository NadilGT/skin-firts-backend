package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"

	"go.mongodb.org/mongo-driver/bson"
)

// DB_FindAdminRole looks up the role of a user in the admin_users collection by Firebase UID.
// Returns the role string and whether the document was found.
func DB_FindAdminRole(firebaseUID string) (string, bool, error) {
	ctx := context.Background()
	var result struct {
		Role string `bson:"role"`
	}
	err := dbConfigs.AdminUserCollection.FindOne(ctx, bson.M{"firebaseUid": firebaseUID}).Decode(&result)
	if err != nil {
		return "", false, nil // not found is not an error — just unknown
	}
	return result.Role, true, nil
}

// DB_FindMobileUserRole looks up the role of a user across the patients and doctor_users
// collections by Firebase UID. Returns the role and whether the document was found.
func DB_FindMobileUserRole(firebaseUID string) (string, bool, error) {
	ctx := context.Background()
	var result struct {
		Role string `bson:"role"`
	}

	// Check patients first
	err := dbConfigs.PatientCollection.FindOne(ctx, bson.M{"firebaseUid": firebaseUID}).Decode(&result)
	if err == nil {
		return result.Role, true, nil
	}

	// Then check doctor_users
	err = dbConfigs.DoctorUserCollection.FindOne(ctx, bson.M{"firebaseUid": firebaseUID}).Decode(&result)
	if err == nil {
		return result.Role, true, nil
	}

	return "", false, nil // not found in either collection
}
