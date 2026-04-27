package api

import (
	"context"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateDoctorWeeklySchedule(c *fiber.Ctx) error {
	var schedule dto.DoctorWeeklySchedule
	if err := c.BodyParser(&schedule); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Resolve branchId
	branchId, err := ResolveBranchId(c, schedule.BranchId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	schedule.BranchId = branchId

	DoctorWeeklyScheduleID, err := dao.GenerateId(context.Background(), "doctorWeeklySchedules", "DWS")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate patient ID",
		})
	}
	schedule.DoctorWeeklyScheduleID = DoctorWeeklyScheduleID
	id, err := dao.DB_CreateDoctorWeeklySchedule(schedule)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create schedule"})
	}
	schedule.ID = id
	return c.Status(fiber.StatusCreated).JSON(schedule)
}

func UpdateDoctorWeeklySchedule(c *fiber.Ctx) error {
	id := c.Query("doctorId")
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var schedule dto.DoctorWeeklySchedule
	if err := c.BodyParser(&schedule); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if err := dao.DB_UpdateDoctorWeeklySchedule(id, branchId, schedule); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Schedule not found for this doctor"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update schedule"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Schedule updated successfully"})
}

func DeleteDoctorWeeklySchedule(c *fiber.Ctx) error {
	id := c.Query("doctorId")
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := dao.DB_DeleteDoctorWeeklySchedule(id, branchId); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Schedule not found for this doctor"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete schedule"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Schedule deleted successfully"})
}

func GetAllDoctorWeeklySchedules(c *fiber.Ctx) error {
	doctorID := c.Query("doctorId")
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	schedules, err := dao.DB_FindAllDoctorWeeklySchedules(doctorID, branchId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch schedules"})
	}
	return c.Status(fiber.StatusOK).JSON(schedules)
}


func GetDoctorAvailableDatesForWeek(c *fiber.Ctx) error {
	doctorID := c.Query("doctorId")
	if doctorID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Doctor ID is required"})
	}

	// 1. Fetch DoctorWeeklySchedule from MongoDB using doctorId, branchId and isActive = true
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	schedules, err := dao.DB_FindAllDoctorWeeklySchedules(doctorID, branchId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch schedules"})
	}

	// 2. Filter active schedules and collect available days
	availableDays := make(map[int]*string) // key: dayOfWeek, value: defaultStartTime
	for _, s := range schedules {
		if s.IsActive {
			for _, day := range s.DaysOfWeek {
				availableDays[day] = s.DefaultStartTime
			}
		}
	}

	// 3. Get current date and calculate the start of the week (Sunday)
	now := time.Now()
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// Sunday is 0. now.Weekday() returns 0-6.
	// To get Sunday of the current week:
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))

	// 4. Loop through the next 14 days (Current Week + Next Week)
	dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	var availableDates []dto.AvailableDate

	for i := 0; i < 14; i++ {
		date := startOfWeek.AddDate(0, 0, i)

		// Skip dates before today
		if date.Before(todayDate) {
			continue
		}

		dayOfWeek := int(date.Weekday())

		// 5. If dayOfWeek exists in schedule.daysOfWeek AND the doctor is actually available on that date
		if startTime, ok := availableDays[dayOfWeek]; ok {
			isAvailable, _, err := dao.DB_CheckDoctorAvailabilityOnDate(doctorID, branchId, date)
			if err == nil && isAvailable {
				availableDates = append(availableDates, dto.AvailableDate{
					Date:             date.Format("2006-01-02"),
					DayOfWeek:        dayOfWeek,
					DayName:          dayNames[dayOfWeek],
					DefaultStartTime: startTime,
				})
			}
		}
	}

	// 6. Return response
	return c.Status(fiber.StatusOK).JSON(dto.AvailableDateResponse{
		AvailableDates: availableDates,
	})
}

func CheckDoctorAvailability(c *fiber.Ctx) error {
	doctorID := c.Query("doctorId")
	dateStr := c.Query("date")
	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if doctorID == "" || dateStr == "" || branchId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Doctor ID, date, and branchId are required"})
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format. Use YYYY-MM-DD"})
	}

	// Check weekly schedule and return a virtual record
	isAvailable, message, err := dao.DB_CheckDoctorAvailabilityOnDate(doctorID, branchId, date)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check weekly schedule"})
	}

	// Try to get defaultStartTime from weekly schedule if available
	var defaultStartTime *string
	dayOfWeek := int(date.Weekday())
	schedules, err := dao.DB_FindAllDoctorWeeklySchedules(doctorID, branchId)
	if err == nil {
		for _, s := range schedules {
			if s.IsActive {
				for _, d := range s.DaysOfWeek {
					if d == dayOfWeek {
						defaultStartTime = s.DefaultStartTime
						break
					}
				}
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"doctorId":           doctorID,
		"date":               dateStr,
		"isAvailable":        isAvailable,
		"estimatedStartTime": defaultStartTime,
		"notes":              message,
	})
}
