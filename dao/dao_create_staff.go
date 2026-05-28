package dao

import (
	"context"
	"errors"
	"fmt"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// DB_CreateStaffUser creates a new staff user in the "staff_users" collection.
// Duplicate check is by email.
func DB_CreateStaffUser(staff dto.StaffUser) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := dbConfigs.StaffUserCollection.CountDocuments(ctx, bson.M{"email": staff.Email})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("a staff user with this email already exists")
	}

	var staffCollection *mongo.Collection = dbConfigs.StaffUserCollection

	_, err = staffCollection.InsertOne(ctx, staff)
	if err != nil {
		return fmt.Errorf("could not insert staff user: %w", err)
	}

	return nil
}
