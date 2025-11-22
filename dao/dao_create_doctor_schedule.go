package dao

import (
	"context"
	"fmt"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DB_CreateOrUpdateDoctorSchedule(schedule dto.DoctorScheduleModel) error {
	filter := bson.M{
		"doctorName": schedule.DoctorName,
		"date":       schedule.Date,
	}

	update := bson.M{
		"$addToSet": bson.M{
			"timeSlots": bson.M{"$each": schedule.TimeSlots}, // APPEND UNIQUE VALUES
		},
		"$set": bson.M{
			"doctorName": schedule.DoctorName,
			"date":       schedule.Date,
			"updatedAt":  time.Now(),
		},
		"$setOnInsert": bson.M{
			"createdAt": time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := dbConfigs.DoctorScheduleCollection.UpdateOne(
		context.Background(),
		filter,
		update,
		opts,
	)

	return err
}

// DB_GetDoctorSchedule retrieves all schedules for a specific doctor
func DB_GetDoctorSchedule(doctorName string) ([]dto.DoctorScheduleModel, error) {
	filter := bson.M{"doctorName": doctorName}

	// Sort by date ascending
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: 1}})

	cursor, err := dbConfigs.DoctorScheduleCollection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var schedules []dto.DoctorScheduleModel
	if err = cursor.All(context.Background(), &schedules); err != nil {
		return nil, err
	}

	return schedules, nil
}

// DB_GetDoctorScheduleByDateRange retrieves schedules for a doctor within a date range
func DB_GetDoctorScheduleByDateRange(doctorName string, startDate, endDate time.Time) ([]dto.DoctorScheduleModel, error) {
	filter := bson.M{
		"doctorName": doctorName,
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "date", Value: 1}})

	cursor, err := dbConfigs.DoctorScheduleCollection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var schedules []dto.DoctorScheduleModel
	if err = cursor.All(context.Background(), &schedules); err != nil {
		return nil, err
	}

	return schedules, nil
}

func DB_DeleteDoctorSchedule(doctorName string, date time.Time) error {
	filter := bson.M{
		"doctorName": doctorName,
		"date":       date,
	}

	_, err := dbConfigs.DoctorScheduleCollection.DeleteOne(context.Background(), filter)
	return err
}

// DB_DeleteTimeSlotFromSchedule removes a specific time slot from a doctor's schedule
func DB_DeleteTimeSlotFromSchedule(doctorName string, date time.Time, timeSlot string) error {
	filter := bson.M{
		"doctorName": doctorName,
		"date":       date,
	}

	update := bson.M{
		"$pull": bson.M{
			"timeSlots": timeSlot,
		},
	}

	result, err := dbConfigs.DoctorScheduleCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("schedule not found for doctor %s on %s", doctorName, date.Format("2006-01-02"))
	}

	return nil
}
