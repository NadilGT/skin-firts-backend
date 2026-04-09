package api

import (
	"fmt"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/functions"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (h *AppointmentStatusHandler) RescheduleAppointment(c *fiber.Ctx) error {
	appointmentID := c.Query("appointmentId")
	if appointmentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Appointment ID is required",
		})
	}

	var req struct {
		AppointmentDate string `json:"appointmentDate"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.AppointmentDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Appointment date is required",
		})
	}

	// Parse the new date
	newDate, err := parseFlexibleDate(req.AppointmentDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date format. Use ISO8601 or YYYY-MM-DD",
		})
	}

	// Prevent past bookings
	if newDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot reschedule appointment to a past date",
		})
	}

	// Get existing appointment to check availability for the same doctor/time
	existingAppointment, err := dao.DB_GetAppointmentByID(appointmentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Appointment not found",
		})
	}

	// Check if the doctor is available on the new date
	isAvailable, reason, err := dao.DB_CheckDoctorAvailabilityOnDate(existingAppointment.DoctorID, newDate)
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

	// Get the next appointment number for the rescheduled date
	nextNum, err := dao.DB_GetNextAppointmentNumber(existingAppointment.DoctorID, newDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate appointment number",
		})
	}

	// Update the appointment
	if err := dao.DB_RescheduleAppointment(appointmentID, newDate, nextNum); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reschedule appointment",
		})
	}

	// Send FCM notification asynchronously
	go h.notifyPatientRescheduled(*existingAppointment, newDate, nextNum)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Appointment rescheduled successfully",
	})
}

// notifyPatientRescheduled sends an FCM push notification to the patient about the rescheduled appointment.
func (h *AppointmentStatusHandler) notifyPatientRescheduled(appointment dto.AppointmentModel, newDate time.Time, nextNum int) {
	if appointment.PatientID == "" {
		log.Printf("⚠️  Notify: appointment %s has no patientId, skipping", appointment.AppointmentID)
		return
	}

	fcmToken, err := dao.DB_GetPatientFCMToken(appointment.PatientID)
	if err != nil {
		log.Printf("⚠️  Notify: could not fetch FCM token for patient %s: %v", appointment.PatientID, err)
		// fcmToken will be empty — SaveAndSendNotification still saves to DB
	}

	dateStr := newDate.Format("Jan 2, 2006")

	title := "Appointment Rescheduled 📅"
	body := fmt.Sprintf("Your appointment with Dr. %s has been rescheduled to %s. Your new queue number is %d.", appointment.DoctorName, dateStr, nextNum)

	data := map[string]string{
		"type":            "APPOINTMENT_RESCHEDULED",
		"appointmentId":   appointment.AppointmentID,
		"status":          "rescheduled",
		"doctorName":      appointment.DoctorName,
		"appointmentDate": newDate.Format("2006-01-02"),
		"queueNumber":     fmt.Sprintf("%d", nextNum),
	}

	// Save to MongoDB first, then fire FCM push (best-effort)
	if err := functions.SaveAndSendNotification(
		h.FirebaseApp,
		fcmToken,
		appointment.PatientID,
		title,
		body,
		"APPOINTMENT_RESCHEDULED",
		data,
	); err != nil {
		log.Printf("⚠️  Notify: pipeline failed for patient %s: %v", appointment.PatientID, err)
	}
}