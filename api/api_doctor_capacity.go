package api

import (
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// ─── GET /doctor-daily-capacity ──────────────────────────────────────────────
// Query params: doctorId, branchId, fromDate (YYYY-MM-DD), toDate (YYYY-MM-DD)
// Returns a list of capacity records, sorted by date ascending.
func GetAllDailyCapacities(c *fiber.Ctx) error {
	doctorID := c.Query("doctorId")
	fromDate := c.Query("fromDate")
	toDate := c.Query("toDate")

	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	caps, err := dao.DB_FindAllDailyCapacities(doctorID, branchId, fromDate, toDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch capacity records"})
	}
	return c.Status(fiber.StatusOK).JSON(caps)
}

// ─── GET /doctor-daily-capacity/single ───────────────────────────────────────
// Query params: doctorId (required), branchId (required), date (required, YYYY-MM-DD)
// Returns a single capacity record, or 404 if not found.
func GetSingleDailyCapacity(c *fiber.Ctx) error {
	doctorID := c.Query("doctorId")
	dateStr := c.Query("date")
	if doctorID == "" || dateStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "doctorId and date are required"})
	}

	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	cap, err := dao.DB_GetDailyCapacity(doctorID, branchId, dateStr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch capacity record"})
	}
	if cap == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "No capacity record found (doctor is unlimited for this date)",
		})
	}
	return c.Status(fiber.StatusOK).JSON(cap)
}

// ─── POST /doctor-daily-capacity ─────────────────────────────────────────────
// Body: { doctorId, branchId, date, max, booked? }
// Admin creates a capacity record manually.
func CreateDailyCapacity(c *fiber.Ctx) error {
	var body dto.DoctorDailyCapacity
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if body.DoctorID == "" || body.Date == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "doctorId and date are required"})
	}
	if body.Max <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "max must be greater than 0"})
	}

	branchId, err := ResolveBranchId(c, body.BranchId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	body.BranchId = branchId

	created, err := dao.DB_CreateDailyCapacity(body)
	if err != nil {
		if err.Error() == "capacity record already exists for this doctor/branch/date" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create capacity record"})
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

// ─── PUT /doctor-daily-capacity ──────────────────────────────────────────────
// Query param: doctorDailyCapacityId (required)
// Body: { max, booked }
// Admin updates max slots or manually adjusts the booked counter using the record's ID.
func UpdateDailyCapacity(c *fiber.Ctx) error {
	capacityId := c.Query("doctorDailyCapacityId")
	if capacityId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "doctorDailyCapacityId is required"})
	}

	var body struct {
		Max    int `json:"max"`
		Booked int `json:"booked"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if body.Max <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "max must be greater than 0"})
	}
	if body.Booked < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "booked cannot be negative"})
	}

	updated, err := dao.DB_UpdateDailyCapacity(capacityId, body.Max, body.Booked)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Capacity record not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update capacity record"})
	}
	return c.Status(fiber.StatusOK).JSON(updated)
}


// ─── DELETE /doctor-daily-capacity ───────────────────────────────────────────
// Query params: doctorId (required), branchId (required), date (required)
// Removes the capacity record — the date becomes unlimited after deletion.
func DeleteDailyCapacity(c *fiber.Ctx) error {
	doctorID := c.Query("doctorId")
	dateStr := c.Query("date")
	if doctorID == "" || dateStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "doctorId and date are required"})
	}

	branchId, err := ResolveBranchId(c, c.Query("branchId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := dao.DB_DeleteDailyCapacity(doctorID, branchId, dateStr); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Capacity record not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete capacity record"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Capacity record deleted. Doctor is now unlimited for this date."})
}
