package api

import (
	"context"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"time"

	"github.com/gofiber/fiber/v2"
)

// --- DoctorWeeklySchedule Handlers ---

func CreateDoctorWeeklySchedule(c *fiber.Ctx) error {
	var schedule dto.DoctorWeeklySchedule
	if err := c.BodyParser(&schedule); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
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
	id := c.Query("id")
	var schedule dto.DoctorWeeklySchedule
	if err := c.BodyParser(&schedule); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if err := dao.DB_UpdateDoctorWeeklySchedule(id, schedule); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update schedule"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Schedule updated successfully"})
}

func DeleteDoctorWeeklySchedule(c *fiber.Ctx) error {
	id := c.Query("id")
	if err := dao.DB_DeleteDoctorWeeklySchedule(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete schedule"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Schedule deleted successfully"})
}

func GetAllDoctorWeeklySchedules(c *fiber.Ctx) error {
	doctorID := c.Query("doctorId")
	schedules, err := dao.DB_FindAllDoctorWeeklySchedules(doctorID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch schedules"})
	}
	return c.Status(fiber.StatusOK).JSON(schedules)
}

// --- DoctorAvailability Handlers ---

func CreateDoctorAvailability(c *fiber.Ctx) error {
	var availability dto.DoctorAvailability
	if err := c.BodyParser(&availability); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	DoctorAvailabilityID, err := dao.GenerateId(context.Background(), "doctorAvailabilities", "DA")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate doctor availability ID",
		})
	}
	availability.DoctorAvailabilityID = DoctorAvailabilityID
	id, err := dao.DB_CreateDoctorAvailability(availability)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create availability"})
	}
	availability.ID = id
	return c.Status(fiber.StatusCreated).JSON(availability)
}

func UpdateDoctorAvailability(c *fiber.Ctx) error {
	id := c.Query("id")
	var availability dto.DoctorAvailability
	if err := c.BodyParser(&availability); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if err := dao.DB_UpdateDoctorAvailability(id, availability); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update availability"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Availability updated successfully"})
}

func DeleteDoctorAvailability(c *fiber.Ctx) error {
	id := c.Query("id")
	if err := dao.DB_DeleteDoctorAvailability(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete availability"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Availability deleted successfully"})
}

func GetAllDoctorAvailabilities(c *fiber.Ctx) error {
	doctorID := c.Query("doctorId")
	availabilities, err := dao.DB_FindAllDoctorAvailabilities(doctorID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch availabilities"})
	}
	return c.Status(fiber.StatusOK).JSON(availabilities)
}

func GetDoctorAvailableDatesForWeek(c *fiber.Ctx) error {
	doctorID := c.Query("doctorId")
	if doctorID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Doctor ID is required"})
	}

	// 1. Fetch DoctorWeeklySchedule from MongoDB using doctorId and isActive = true
	schedules, err := dao.DB_FindAllDoctorWeeklySchedules(doctorID)
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
	// Sunday is 0. now.Weekday() returns 0-6.
	// To get Sunday of the current week:
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))

	// 4. Loop through the next 7 days (Sunday → Saturday)
	dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	var availableDates []dto.AvailableDate

	for i := 0; i < 7; i++ {
		date := startOfWeek.AddDate(0, 0, i)
		dayOfWeek := int(date.Weekday())

		// 5. If dayOfWeek exists in schedule.daysOfWeek → include it
		if startTime, ok := availableDays[dayOfWeek]; ok {
			availableDates = append(availableDates, dto.AvailableDate{
				Date:             date.Format("2006-01-02"),
				DayOfWeek:        dayOfWeek,
				DayName:          dayNames[dayOfWeek],
				DefaultStartTime: startTime,
			})
		}
	}

	// 6. Return response
	return c.Status(fiber.StatusOK).JSON(dto.AvailableDateResponse{
		AvailableDates: availableDates,
	})
}
