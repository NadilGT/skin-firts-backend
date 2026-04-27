package dao

import (
	"context"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// --- DoctorWeeklySchedule DAO ---

func DB_CreateDoctorWeeklySchedule(schedule dto.DoctorWeeklySchedule) (string, error) {
	result, err := dbConfigs.DoctorWeeklyScheduleCollection.InsertOne(context.Background(), schedule)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func DB_UpdateDoctorWeeklySchedule(id string, branchId string, schedule dto.DoctorWeeklySchedule) error {
	filter := bson.M{"doctorId": id, "branchId": branchId}
	update := bson.M{
		"$set": bson.M{
			"daysOfWeek":       schedule.DaysOfWeek,
			"defaultStartTime": schedule.DefaultStartTime,
			"isActive":         schedule.IsActive,
		},
	}
	result, err := dbConfigs.DoctorWeeklyScheduleCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func DB_DeleteDoctorWeeklySchedule(id string, branchId string) error {
	result, err := dbConfigs.DoctorWeeklyScheduleCollection.DeleteOne(context.Background(), bson.M{"doctorId": id, "branchId": branchId})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func DB_FindAllDoctorWeeklySchedules(doctorID string, branchId string) ([]dto.DoctorWeeklySchedule, error) {
	filter := bson.M{}
	if doctorID != "" {
		filter["doctorId"] = doctorID
	}
	if branchId != "" {
		filter["branchId"] = branchId
	}
	cursor, err := dbConfigs.DoctorWeeklyScheduleCollection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var schedules []dto.DoctorWeeklySchedule
	if err = cursor.All(context.Background(), &schedules); err != nil {
		return nil, err
	}
	return schedules, nil
}

// --- DoctorAvailability DAO ---

func DB_CheckDoctorAvailabilityOnDate(doctorID string, branchId string, date time.Time) (bool, string, error) {
	dayOfWeek := int(date.Weekday())

	// Check weekly schedule
	cursor, err := dbConfigs.DoctorWeeklyScheduleCollection.Find(context.Background(), bson.M{
		"doctorId": doctorID,
		"branchId": branchId,
		"isActive": true,
	})
	if err != nil {
		return false, "Failed to fetch doctor schedule", err
	}
	defer cursor.Close(context.Background())

	var schedules []dto.DoctorWeeklySchedule
	if err = cursor.All(context.Background(), &schedules); err != nil {
		return false, "Failed to decode doctor schedule", err
	}

	for _, s := range schedules {
		for _, day := range s.DaysOfWeek {
			if day == dayOfWeek {
				return true, "", nil
			}
		}
	}

	return false, "Doctor does not have a schedule for this day of the week", nil
}
