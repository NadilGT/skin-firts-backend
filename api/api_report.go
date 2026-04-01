package api

import (
	"context"
	"fmt"
	"io"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dto"
	"lawyerSL-Backend/functions"
	"path/filepath"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReportHandler struct {
	App *firebase.App
}

func NewReportHandler(app *firebase.App) *ReportHandler {
	return &ReportHandler{App: app}
}

// UploadReport handles POST /api/reports/upload
func (h *ReportHandler) UploadReport(c *fiber.Ctx) error {
	// 1. Parse Multipart Form
	fileHeader, err := c.FormFile("report")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to retrieve report file from form payload: " + err.Error(),
		})
	}

	patientId := c.FormValue("patientId")
	appointmentId := c.FormValue("appointmentId")
	title := c.FormValue("title")
	description := c.FormValue("description")
	uploadedBy := c.FormValue("uploadedBy")

	if patientId == "" || appointmentId == "" || title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields: patientId, appointmentId, and title are mandatory",
		})
	}

	// 2. Upload to Firebase Storage
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open uploaded file",
		})
	}
	defer file.Close()

	ctx := context.Background()
	client, err := h.App.Storage(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to initialize Firebase Storage client",
		})
	}

	bucket, err := client.DefaultBucket()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to resolve default storage bucket. Check your FIREBASE_STORAGE_BUCKET env variable.",
		})
	}

	// Generate a unique filename in the 'reports' folder
	filename := fmt.Sprintf("reports/%d_%s", time.Now().UnixNano(), fileHeader.Filename)
	obj := bucket.Object(filename)
	writer := obj.NewWriter(ctx)
	
	// Set the file to be publicly accessible
	writer.ObjectAttrs.PredefinedACL = "publicRead"

	if _, err := io.Copy(writer, file); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to transmit file to cloud storage: " + err.Error(),
		})
	}
	
	if err := writer.Close(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to properly conclude storage upload: " + err.Error(),
		})
	}

	// Construct the public GCS URL
	bucketAttrs, err := bucket.Attrs(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to extract bucket metadata: " + err.Error(),
		})
	}
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketAttrs.Name, filename)

	// 3. Generate a human-readable Report ID (e.g., REP-001)
	reportId, err := dao.GenerateId(ctx, "reports", "REP")
	if err != nil {
		// Fallback if ID generator fails, though it shouldn't
		reportId = "REP-" + fmt.Sprintf("%d", time.Now().Unix())
	}

	// 4. Create and Save the Report Model
	report := dto.ReportModel{
		ID:            primitive.NewObjectID(),
		ReportID:      reportId,
		PatientID:     patientId,
		AppointmentID: appointmentId,
		Title:         title,
		Description:   description,
		FileURL:       publicURL,
		FileType:      filepath.Ext(fileHeader.Filename),
		UploadedBy:    uploadedBy,
		Status:        "AVAILABLE",
		CreatedAt:     time.Now(),
	}

	if err := dao.DB_SaveReport(report); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save report details to database: " + err.Error(),
		})
	}

	// 5. Send FCM Notification to the Patient
	fcmToken, err := dao.DB_GetPatientFCMToken(patientId)
	if err == nil && fcmToken != "" {
		notificationTitle := "Report Ready 📄"
		notificationBody := fmt.Sprintf("Your report for '%s' is now available", title)
		data := map[string]string{
			"type":      "REPORT_READY",
			"reportId":  reportId,
			"patientId": patientId,
		}
		
		// Send notification asynchronously in a goroutine to not block the response
		go func() {
			if err := functions.SendFCMNotification(h.App, fcmToken, notificationTitle, notificationBody, data); err != nil {
				fmt.Printf("⚠️  Failed to send FCM notification: %v\n", err)
			}
		}()
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Report uploaded and patient notified successfully",
		"reportId":  reportId,
		"fileUrl":   publicURL,
		"patientId": patientId,
	})
}
