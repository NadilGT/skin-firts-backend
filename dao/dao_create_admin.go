package dao

import (
	"context"
	"errors"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"

	"go.mongodb.org/mongo-driver/bson"
)

// DB_CreateAdminUser inserts a new admin into the "admin_users" collection.
// Duplicate check is by email (firebaseUid is optional/legacy).
func DB_CreateAdminUser(admin dto.AdminUser) error {
	ctx := context.Background()

	count, err := dbConfigs.AdminUserCollection.CountDocuments(ctx, bson.M{"email": admin.Email})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("an admin with this email already exists")
	}

	_, err = dbConfigs.AdminUserCollection.InsertOne(ctx, admin)
	return err
}
