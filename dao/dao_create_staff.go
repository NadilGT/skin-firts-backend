package dao

import (
	"context"
	"fmt"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// DB_CreateStaffUser creates a new staff user in the "staff_users" collection
func DB_CreateStaffUser(staff dto.StaffUser) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var staffCollection *mongo.Collection = dbConfigs.StaffUserCollection

	_, err := staffCollection.InsertOne(ctx, staff)
	if err != nil {
		return fmt.Errorf("could not insert staff user: %w", err)
	}

	return nil
}
