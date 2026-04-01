package api

import (
	"fmt"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/functions"
	"log"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gofiber/fiber/v2"
)

// AppointmentStatusHandler holds the Firebase app so it can send FCM notifications.
type AppointmentStatusHandler struct {
	FirebaseApp *firebase.App
}

// NewAppointmentStatusHandler creates a handler with the shared Firebase app.
func NewAppointmentStatusHandler(firebaseApp *firebase.App) *AppointmentStatusHandler {
	return &AppointmentStatusHandler{FirebaseApp: firebaseApp}
}

// UpdateAppointmentStatus handles PATCH /appointments/id/status
// It updates the appointment status in MongoDB and sends an FCM push notification to the patient.
func (h *AppointmentStatusHandler) UpdateAppointmentStatus(c *fiber.Ctx) error {
	idParam := c.Query("appointmentId")
	if idParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Appointment ID is required",
		})
	}

	var req dto.UpdateAppointmentStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// 1. Update status in DB
	if err := dao.DB_UpdateAppointmentStatus(idParam, req.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update appointment status",
		})
	}

	// 2. Fire-and-forget: send FCM notification to the patient
	go h.notifyPatient(idParam, req.Status)

	return c.JSON(fiber.Map{
		"message":   "Appointment status updated successfully",
		"status":    req.Status,
		"updatedAt": time.Now(),
	})
}

// notifyPatient fetches appointment + patient FCM token and sends the push notification.
// Runs in a goroutine so it never delays the HTTP response.
func (h *AppointmentStatusHandler) notifyPatient(appointmentID string, status string) {
	// Fetch the full appointment to get patientId and appointment details
	appointment, err := dao.DB_GetAppointmentByAppointmentID(appointmentID)
	if err != nil {
		log.Printf("⚠️  FCM notify: could not fetch appointment %s: %v", appointmentID, err)
		return
	}

	if appointment.PatientID == "" {
		log.Printf("⚠️  FCM notify: appointment %s has no patientId, skipping", appointmentID)
		return
	}

	// Fetch FCM token from patients collection (keyed by Firebase UID = patientId)
	fcmToken, err := dao.DB_GetPatientFCMToken(appointment.PatientID)
	if err != nil {
		log.Printf("⚠️  FCM notify: could not fetch FCM token for patient %s: %v", appointment.PatientID, err)
		return
	}

	title, body := buildNotificationContent(status, appointment)

	data := map[string]string{
		"appointmentId":   appointment.AppointmentID,
		"status":          status,
		"doctorName":      appointment.DoctorName,
		"appointmentDate": appointment.AppointmentDate.Format("2006-01-02"),
	}

	if err := functions.SendFCMNotification(h.FirebaseApp, fcmToken, title, body, data); err != nil {
		log.Printf("⚠️  FCM notify: send failed for patient %s: %v", appointment.PatientID, err)
	}
}

// buildNotificationContent returns a human-friendly title and body for each status.
func buildNotificationContent(status string, a dto.AppointmentModel) (title, body string) {
	dateStr := a.AppointmentDate.Format("Jan 2, 2006")
	switch status {
	case "confirmed":
		title = "Appointment Confirmed ✅"
		body = fmt.Sprintf("Your appointment with %s on %s has been confirmed.", a.DoctorName, dateStr)
	case "cancelled":
		title = "Appointment Cancelled ❌"
		body = fmt.Sprintf("Your appointment with %s on %s has been cancelled.", a.DoctorName, dateStr)
	case "completed":
		title = "Appointment Completed 🎉"
		body = fmt.Sprintf("Your appointment with %s on %s is now marked as completed.", a.DoctorName, dateStr)
	case "running":
		title = "Your Turn is Coming 🏥"
		body = fmt.Sprintf("Dr. %s is now seeing patients. Please be ready.", a.DoctorName)
	case "pending":
		title = "Appointment Pending ⏳"
		body = fmt.Sprintf("Your appointment with %s on %s is now pending.", a.DoctorName, dateStr)
	default:
		title = "Appointment Update"
		body = fmt.Sprintf("Your appointment status has been updated to: %s", status)
	}
	return
}
