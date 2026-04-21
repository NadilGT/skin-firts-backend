package api

import (
	"context"
	"fmt"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

// getBranchIdFromContext resolves branchId based on the caller's role:
//   - PATIENT & SUPER_ADMIN → from request body (allows booking for any branch)
//   - All others (admin, doctor, staff) → from JWT token claim
func getBranchIdFromContext(c *fiber.Ctx, reqBranchId string) (string, error) {
	role, _ := c.Locals("role").(string)

	// Patients and Super Admins can specify any branch in the request body
	if role == "patient" || role == "super_admin" {
		if reqBranchId == "" {
			return "", fmt.Errorf("branchId is required")
		}
		return reqBranchId, nil
	}

	// Regular Admin / Doctor / Staff — fixed branch from JWT
	branchId, _ := c.Locals("branchId").(string)
	if branchId == "" {
		return "", fmt.Errorf("no branchId found in token; please contact your administrator")
	}
	return branchId, nil
}

var dateFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,       // 2025-11-18T10:30:00Z
	"2006-01-02",       // 2025-11-18
	"2006-01-02 15:04", // optional
	"2006-01-02 15:04:05",
}

func parseFlexibleDate(dateStr string) (time.Time, error) {
	for _, format := range dateFormats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date")
}

func CreateAppointment(c *fiber.Ctx) error {
	var req dto.CreateAppointmentRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// ── 1. Resolve branchId from context (role-based) ──────────────────────────
	// PATIENT  → must send branchId in request body (mobile app selects a branch)
	// Others   → branchId is fixed in their JWT token (admin/doctor/staff)
	branchId, err := getBranchIdFromContext(c, req.BranchId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// ── 2. Validate branch exists ───────────────────────────────────────────────
	_, err = dao.DB_GetBranchByBranchId(branchId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Branch not found: " + branchId,
		})
	}

	// ── 3. Validate doctor is assigned to this branch ───────────────────────────
	doctor, err := dao.DB_GetDoctorInfoByDoctorId(req.DoctorID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Doctor not found: " + req.DoctorID,
		})
	}

	doctorInBranch := false
	for _, bid := range doctor.BranchIds {
		if bid == branchId {
			doctorInBranch = true
			break
		}
	}
	if !doctorInBranch {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": fmt.Sprintf("Doctor %s is not assigned to branch %s", req.DoctorID, branchId),
		})
	}

	// ── 4. Generate appointment ID ──────────────────────────────────────────────
	id, err := dao.GenerateId(context.Background(), "appointments", "APP")
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	req.AppointmentID = id

	fmt.Println("Incoming appointmentDate:", req.AppointmentDate)

	// ── 5. Parse flexible date input ────────────────────────────────────────────
	appointmentDate, err := parseFlexibleDate(req.AppointmentDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date format. Use ISO8601 or YYYY-MM-DD",
		})
	}

	// Prevent past bookings
	if appointmentDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot book appointment in the past",
		})
	}

	// ── 6. Check doctor availability/schedule ───────────────────────────────────
	isAvailable, reason, err := dao.DB_CheckDoctorAvailabilityOnDate(req.DoctorID, branchId, appointmentDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to check doctor availability",
			"details": err.Error(),
		})
	}
	if !isAvailable {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": reason,
		})
	}

	nextNum, err := dao.DB_GetNextAppointmentNumber(req.DoctorID, branchId, appointmentDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate appointment number",
		})
	}

	// ── 7. Build and persist the appointment ────────────────────────────────────
	appointment := dto.AppointmentModel{
		AppointmentID:     req.AppointmentID,
		AppointmentNumber: nextNum,
		PatientID:         req.PatientID,
		PatientName:       req.PatientName,
		PatientEmail:      req.PatientEmail,
		PatientPhone:      req.PatientPhone,
		DoctorID:          req.DoctorID,
		DoctorName:        req.DoctorName,
		DoctorSpecialty:   req.DoctorSpecialty,
		AppointmentDate:   appointmentDate,
		Status:            "pending",
		Notes:             req.Notes,
		BranchId:          branchId, // ← enforced by backend, never trusted from frontend
	}

	if err := dao.DB_CreateAppointment(appointment); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create appointment",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":                 "Appointment booked successfully",
		"appointment":             appointment,
		"next_appointment_number": nextNum,
	})
}

