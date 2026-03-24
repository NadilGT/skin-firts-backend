package dao

import (
	"context"
	"errors"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
)

// DB_CreateAdminUser inserts a new admin into the "admin_users" collection.
// It returns an error if a document with the same FirebaseUID already exists.
func DB_CreateAdminUser(admin dto.AdminUser) error {
	ctx := context.Background()

	// Prevent duplicate Firebase UID
	count, err := dbConfigs.AdminUserCollection.CountDocuments(ctx, bson.M{"firebaseUid": admin.FirebaseUID})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("an admin with this Firebase UID already exists")
	}

	_, err = dbConfigs.AdminUserCollection.InsertOne(ctx, admin)
	return err
}
