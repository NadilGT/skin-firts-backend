package dao

import (
	"context"
	"fmt"
	"time"

	"lawyerSL-Backend/dbConfigs"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// collectionByName maps a collection name string to the live *mongo.Collection.
func collectionByName(name string) *mongo.Collection {
	switch name {
	case "admin_users":
		return dbConfigs.AdminUserCollection
	case "doctor_users":
		return dbConfigs.DoctorUserCollection
	case "staff_users":
		return dbConfigs.StaffUserCollection
	case "patients":
		return dbConfigs.PatientCollection
	default:
		return nil
	}
}

// ---------------------------------------------------------------------------
// DB_SetPasswordHash — update the passwordHash field for a user by email
// ---------------------------------------------------------------------------

// DB_SetPasswordHash updates the passwordHash field for the user with the given
// email in the specified collection.  Collection must be one of the four user
// collection names.
func DB_SetPasswordHash(email, collection, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := collectionByName(collection)
	if col == nil {
		return fmt.Errorf("unknown collection: %s", collection)
	}

	result, err := col.UpdateOne(
		ctx,
		bson.M{"email": email},
		bson.M{"$set": bson.M{"passwordHash": hash}},
		options.Update().SetUpsert(false),
	)
	if err != nil {
		return fmt.Errorf("failed to update password hash: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("no user found with email %s in %s", email, collection)
	}

	return nil
}

// ---------------------------------------------------------------------------
// DB_ClearMustChangePassword — remove the mustChangePassword flag
// ---------------------------------------------------------------------------

// DB_ClearMustChangePassword sets mustChangePassword = false after the user
// completes their first-time password change.
func DB_ClearMustChangePassword(email, collection string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := collectionByName(collection)
	if col == nil {
		return fmt.Errorf("unknown collection: %s", collection)
	}

	_, err := col.UpdateOne(
		ctx,
		bson.M{"email": email},
		bson.M{"$set": bson.M{"mustChangePassword": false}},
	)
	return err
}

// ---------------------------------------------------------------------------
// DB_UpdateUserRoleAndBranch — update role + branchId for a user by email
// ---------------------------------------------------------------------------

// DB_UpdateUserRoleAndBranch updates the role and branchId for a user across
// whichever collection they belong to.  Searches all 4 collections.
func DB_UpdateUserRoleAndBranch(email, role, branchId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collections := []struct {
		col  *mongo.Collection
		name string
	}{
		{dbConfigs.AdminUserCollection, "admin_users"},
		{dbConfigs.DoctorUserCollection, "doctor_users"},
		{dbConfigs.StaffUserCollection, "staff_users"},
		{dbConfigs.PatientCollection, "patients"},
	}

	update := bson.M{"$set": bson.M{"role": role, "branchId": branchId}}

	for _, entry := range collections {
		result, err := entry.col.UpdateOne(ctx, bson.M{"email": email}, update)
		if err != nil {
			return fmt.Errorf("DB error on %s: %w", entry.name, err)
		}
		if result.MatchedCount > 0 {
			return nil
		}
	}

	return fmt.Errorf("user not found with email %s in any collection", email)
}

// ---------------------------------------------------------------------------
// DB_UpdateUserStatus — update account status (ACTIVE / INACTIVE / SUSPENDED)
// ---------------------------------------------------------------------------

// DB_UpdateUserStatus sets the status field for a user located by email.
func DB_UpdateUserStatus(email, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collections := []*mongo.Collection{
		dbConfigs.AdminUserCollection,
		dbConfigs.DoctorUserCollection,
		dbConfigs.StaffUserCollection,
		dbConfigs.PatientCollection,
	}

	for _, col := range collections {
		result, err := col.UpdateOne(ctx, bson.M{"email": email},
			bson.M{"$set": bson.M{"status": status}})
		if err != nil {
			return err
		}
		if result.MatchedCount > 0 {
			return nil
		}
	}

	return fmt.Errorf("user not found with email %s", email)
}

// ---------------------------------------------------------------------------
// DB_ListAllUsers — aggregate all user collections for admin list view
// ---------------------------------------------------------------------------

// RawUserListing is a minimal view used for the admin user list endpoint.
type RawUserListing struct {
	UserID    string `bson:"userId"    json:"userId"`
	Name      string `bson:"name"      json:"name"`
	Email     string `bson:"email"     json:"email"`
	Role      string `bson:"role"      json:"role"`
	BranchId  string `bson:"branchId"  json:"branchId"`
	Status    string `bson:"status"    json:"status"`
	CreatedAt string `bson:"createdAt" json:"createdAt"`
}

// DB_ListAllUsers returns a combined user list from all 4 collections.
// PasswordHash is never included.
func DB_ListAllUsers() ([]RawUserListing, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	projection := options.Find().SetProjection(bson.M{
		"passwordHash": 0, // explicitly exclude password
		"firebaseUid":  0,
	})

	collections := []struct {
		col  *mongo.Collection
		name string
	}{
		{dbConfigs.AdminUserCollection, "admin_users"},
		{dbConfigs.DoctorUserCollection, "doctor_users"},
		{dbConfigs.StaffUserCollection, "staff_users"},
		{dbConfigs.PatientCollection, "patients"},
	}

	var all []RawUserListing
	for _, entry := range collections {
		cursor, err := entry.col.Find(ctx, bson.M{}, projection)
		if err != nil {
			return nil, fmt.Errorf("failed to query %s: %w", entry.name, err)
		}
		var rows []RawUserListing
		if err := cursor.All(ctx, &rows); err != nil {
			return nil, err
		}
		all = append(all, rows...)
	}

	return all, nil
}
