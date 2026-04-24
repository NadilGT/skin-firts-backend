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

func DB_CreateDoctorAvailability(availability dto.DoctorAvailability) (string, error) {
	availability.CreatedAt = time.Now()
	availability.UpdatedAt = time.Now()
	result, err := dbConfigs.DoctorAvailabilityCollection.InsertOne(context.Background(), availability)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func DB_UpdateDoctorAvailability(id string, branchId string, availability dto.DoctorAvailability) error {
	filter := bson.M{"doctorAvailabilityId": id}
	if branchId != "" {
		filter["branchId"] = branchId
	}
	update := bson.M{
		"$set": bson.M{
			"doctorId":           availability.DoctorID,
			"date":               availability.Date,
			"isAvailable":        availability.IsAvailable,
			"estimatedStartTime": availability.EstimatedStartTime,
			"maxPatients":        availability.MaxPatients,
			"notes":              availability.Notes,
			"updatedAt":          time.Now(),
		},
	}
	result, err := dbConfigs.DoctorAvailabilityCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func DB_DeleteDoctorAvailability(id string, branchId string) error {
	filter := bson.M{"doctorAvailabilityId": id}
	if branchId != "" {
		filter["branchId"] = branchId
	}
	result, err := dbConfigs.DoctorAvailabilityCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func DB_FindAllDoctorAvailabilities(doctorID string, branchId string) ([]dto.DoctorAvailability, error) {
	filter := bson.M{}
	if doctorID != "" {
		filter["doctorId"] = doctorID
	}
	if branchId != "" {
		filter["branchId"] = branchId
	}
	cursor, err := dbConfigs.DoctorAvailabilityCollection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var availabilities []dto.DoctorAvailability
	if err = cursor.All(context.Background(), &availabilities); err != nil {
		return nil, err
	}
	return availabilities, nil
}

func DB_CheckDoctorAvailabilityOnDate(doctorID string, branchId string, date time.Time) (bool, string, error) {
	dateStr := date.Format("2006-01-02")
	dayOfWeek := int(date.Weekday())

	// 1. Check specific availability override
	var availability dto.DoctorAvailability
	err := dbConfigs.DoctorAvailabilityCollection.FindOne(context.Background(), bson.M{
		"doctorId": doctorID,
		"branchId": branchId,
		"date":     dateStr,
	}).Decode(&availability)

	if err == nil {
		if !availability.IsAvailable {
			return false, "Doctor is marked as unavailable on this date", nil
		}
		// If doctor is available via override, check max patients if specified
		if availability.MaxPatients != nil {
			startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
			endOfDay := startOfDay.Add(24 * time.Hour)
			count, err := dbConfigs.AppointmentCollection.CountDocuments(context.Background(), bson.M{
				"doctorId": doctorID,
				"branchId": branchId,
				"appointmentDate": bson.M{
					"$gte": startOfDay,
					"$lt":  endOfDay,
				},
				"status": bson.M{"$ne": "cancelled"},
			})
			if err == nil && int(count) >= *availability.MaxPatients {
				return false, "Doctor has reached the maximum number of patients for this date", nil
			}
		}
		return true, "", nil
	}

	// 2. No specific override, check weekly schedule
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

func DB_FindDoctorAvailabilityByDate(doctorID string, branchId string, date string) (*dto.DoctorAvailability, error) {
	var availability dto.DoctorAvailability
	err := dbConfigs.DoctorAvailabilityCollection.FindOne(context.Background(), bson.M{
		"doctorId": doctorID,
		"branchId": branchId,
		"date":     date,
	}).Decode(&availability)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &availability, nil
}
