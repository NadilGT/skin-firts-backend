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
	"go.mongodb.org/mongo-driver/mongo/options"
)

// normalizeDate strips time from a date and returns UTC start-of-day string "YYYY-MM-DD".
func normalizeDate(t time.Time) string {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).Format("2006-01-02")
}

// DB_EnsureCapacity lazily creates a DoctorDailyCapacity record when one does
// not yet exist for (doctorId, branchId, date).
//
//   - If the schedule has MaxPatients == 0, capacity is treated as unlimited
//     and no document is created (BookAppointment will short-circuit).
//
// It uses an upsert with $setOnInsert so concurrent requests produce exactly
// one document (no race condition, no duplicate-key panic).
func DB_EnsureCapacity(doctorID, branchId string, date time.Time) error {
	dateStr := normalizeDate(date)
	dayOfWeek := int(date.Weekday())

	// 1. Look up the active weekly schedule for this doctor/branch/weekday.
	cursor, err := dbConfigs.DoctorWeeklyScheduleCollection.Find(context.Background(), bson.M{
		"doctorId": doctorID,
		"branchId": branchId,
		"isActive": true,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch weekly schedule: %w", err)
	}
	defer cursor.Close(context.Background())

	var schedules []dto.DoctorWeeklySchedule
	if err = cursor.All(context.Background(), &schedules); err != nil {
		return fmt.Errorf("failed to decode weekly schedule: %w", err)
	}

	// Find the matching day.
	maxPatients := 0
	found := false
	for _, s := range schedules {
		for _, d := range s.DaysOfWeek {
			if d == dayOfWeek {
				maxPatients = s.MaxPatients
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("doctor does not have a schedule for this day of the week")
	}

	// MaxPatients == 0 means unlimited — no capacity document needed.
	if maxPatients == 0 {
		return nil
	}

	// 2. Generate a new ID only for insertion ($setOnInsert will ignore it if doc already exists).
	genId, err := GenerateId(context.Background(), "doctorDailyCapacity", "DDC")
	if err != nil {
		return fmt.Errorf("failed to generate capacity ID: %w", err)
	}

	// 3. Upsert: only set fields on insert to avoid overwriting existing counters.
	filter := bson.M{
		"doctorId": doctorID,
		"branchId": branchId,
		"date":     dateStr,
	}
	update := bson.M{
		"$setOnInsert": bson.M{
			"doctorDailyCapacityId": genId,
			"doctorId":             doctorID,
			"branchId":             branchId,
			"date":                 dateStr,
			"booked":               0,
			"max":                  maxPatients,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = dbConfigs.DoctorDailyCapacityCollection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert capacity document: %w", err)
	}
	return nil
}

// DB_BookAppointmentCapacity atomically increments the booked counter only
// when booked < max, preventing overbooking even under concurrent load.
//
//   - If no capacity document exists (MaxPatients == 0 → unlimited), the
//     function returns nil immediately (booking always allowed).
//
// forceBook = true lets admins bypass the capacity check (optional override).
func DB_BookAppointmentCapacity(doctorID, branchId string, date time.Time, forceBook bool) error {
	dateStr := normalizeDate(date)

	// Check if a capacity document exists at all.
	var cap dto.DoctorDailyCapacity
	err := dbConfigs.DoctorDailyCapacityCollection.FindOne(context.Background(), bson.M{
		"doctorId": doctorID,
		"branchId": branchId,
		"date":     dateStr,
	}).Decode(&cap)

	if err == mongo.ErrNoDocuments {
		// No capacity document → unlimited (MaxPatients == 0). Allow booking.
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to fetch capacity: %w", err)
	}

	// Admin force-book bypasses the capacity limit.
	if forceBook {
		_, err = dbConfigs.DoctorDailyCapacityCollection.UpdateOne(
			context.Background(),
			bson.M{"doctorId": doctorID, "branchId": branchId, "date": dateStr},
			bson.M{"$inc": bson.M{"booked": 1}},
		)
		return err
	}

	// Atomic: only update if booked < max.
	filter := bson.M{
		"doctorId":         doctorID,
		"branchId":         branchId,
		"date":             dateStr,
		"$expr":            bson.M{"$lt": bson.A{"$booked", "$max"}},
	}
	result, err := dbConfigs.DoctorDailyCapacityCollection.UpdateOne(
		context.Background(),
		filter,
		bson.M{"$inc": bson.M{"booked": 1}},
	)
	if err != nil {
		return fmt.Errorf("failed to increment booking count: %w", err)
	}
	if result.MatchedCount == 0 {
		return errors.New("doctor is fully booked for this day")
	}
	return nil
}

// DB_ReleaseAppointmentCapacity decrements the booked counter when an
// appointment is cancelled or rescheduled away from a date.
// Safe to call even when no capacity document exists (unlimited schedules).
func DB_ReleaseAppointmentCapacity(doctorID, branchId string, date time.Time) error {
	dateStr := normalizeDate(date)

	_, err := dbConfigs.DoctorDailyCapacityCollection.UpdateOne(
		context.Background(),
		bson.M{
			"doctorId": doctorID,
			"branchId": branchId,
			"date":     dateStr,
			"booked":   bson.M{"$gt": 0}, // never go below zero
		},
		bson.M{"$inc": bson.M{"booked": -1}},
	)
	return err
}

// DB_GetDailyCapacity returns the current capacity snapshot for a given
// doctor/branch/date. Returns nil if no document exists (unlimited).
func DB_GetDailyCapacity(doctorID, branchId, dateStr string) (*dto.DoctorDailyCapacity, error) {
	var cap dto.DoctorDailyCapacity
	err := dbConfigs.DoctorDailyCapacityCollection.FindOne(context.Background(), bson.M{
		"doctorId": doctorID,
		"branchId": branchId,
		"date":     dateStr,
	}).Decode(&cap)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cap, nil
}

// ─── Admin CRUD ──────────────────────────────────────────────────────────────

// DB_CreateDailyCapacity lets an admin manually create a capacity record for a
// specific doctor / branch / date without waiting for the first booking to
// lazily initialise it.  Returns a conflict error if the record already exists.
func DB_CreateDailyCapacity(cap dto.DoctorDailyCapacity) (*dto.DoctorDailyCapacity, error) {
	cap.Date = normalizeDate(mustParseDate(cap.Date))

	// Generate a unique ID for this record.
	genId, err := GenerateId(context.Background(), "doctorDailyCapacity", "DDC")
	if err != nil {
		return nil, fmt.Errorf("failed to generate capacity ID: %w", err)
	}
	cap.DoctorDailyCapacityId = genId

	filter := bson.M{"doctorId": cap.DoctorID, "branchId": cap.BranchId, "date": cap.Date}
	update := bson.M{
		"$setOnInsert": bson.M{
			"doctorDailyCapacityId": cap.DoctorDailyCapacityId,
			"doctorId":             cap.DoctorID,
			"branchId":             cap.BranchId,
			"date":                 cap.Date,
			"booked":               cap.Booked,
			"max":                  cap.Max,
		},
	}
	opts := options.Update().SetUpsert(true)
	result, err := dbConfigs.DoctorDailyCapacityCollection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		return nil, err
	}
	if result.UpsertedCount == 0 {
		return nil, errors.New("capacity record already exists for this doctor/branch/date")
	}
	return &cap, nil
}

// DB_UpdateDailyCapacity lets an admin update the max limit and/or the booked
// count for an existing capacity record, identified by its doctorDailyCapacityId.
func DB_UpdateDailyCapacity(capacityId string, max, booked int) (*dto.DoctorDailyCapacity, error) {
	filter := bson.M{"doctorDailyCapacityId": capacityId}
	update := bson.M{
		"$set": bson.M{
			"max":    max,
			"booked": booked,
		},
	}
	result, err := dbConfigs.DoctorDailyCapacityCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return nil, err
	}
	if result.MatchedCount == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return DB_GetDailyCapacityById(capacityId)
}

// DB_GetDailyCapacityById fetches a capacity record by its doctorDailyCapacityId.
func DB_GetDailyCapacityById(capacityId string) (*dto.DoctorDailyCapacity, error) {
	var cap dto.DoctorDailyCapacity
	err := dbConfigs.DoctorDailyCapacityCollection.FindOne(context.Background(), bson.M{
		"doctorDailyCapacityId": capacityId,
	}).Decode(&cap)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cap, nil
}

// DB_DeleteDailyCapacity removes a capacity record entirely.  After deletion
// the doctor/date will be treated as unlimited until a new record is created.
func DB_DeleteDailyCapacity(doctorID, branchId, dateStr string) error {
	dateStr = normalizeDate(mustParseDate(dateStr))

	result, err := dbConfigs.DoctorDailyCapacityCollection.DeleteOne(context.Background(), bson.M{
		"doctorId": doctorID,
		"branchId": branchId,
		"date":     dateStr,
	})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// DB_FindAllDailyCapacities lists capacity records filtered by optional
// doctorId, branchId and/or date range (fromDate, toDate both inclusive,
// format "YYYY-MM-DD").  Pass empty strings to skip a filter.
func DB_FindAllDailyCapacities(doctorID, branchId, fromDate, toDate string) ([]dto.DoctorDailyCapacity, error) {
	filter := bson.M{}
	if doctorID != "" {
		filter["doctorId"] = doctorID
	}
	if branchId != "" {
		filter["branchId"] = branchId
	}
	// Date range filtering
	if fromDate != "" || toDate != "" {
		dateFilter := bson.M{}
		if fromDate != "" {
			dateFilter["$gte"] = normalizeDate(mustParseDate(fromDate))
		}
		if toDate != "" {
			dateFilter["$lte"] = normalizeDate(mustParseDate(toDate))
		}
		filter["date"] = dateFilter
	}

	cursor, err := dbConfigs.DoctorDailyCapacityCollection.Find(
		context.Background(),
		filter,
		options.Find().SetSort(bson.D{{Key: "date", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var caps []dto.DoctorDailyCapacity
	if err = cursor.All(context.Background(), &caps); err != nil {
		return nil, err
	}
	return caps, nil
}

// mustParseDate parses "YYYY-MM-DD" and returns zero time on failure (safe for
// normalizeDate which only needs year/month/day).
func mustParseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

