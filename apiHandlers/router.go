package apiHandlers

import (
	"lawyerSL-Backend/api"

	firebase "firebase.google.com/go/v4"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, authMiddleware *AuthMiddleware, firebaseApp *firebase.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello from Fiber on Render!")
	})

	// ========== APP DOWNLOAD ROUTE ==========
	app.Get("/download/app", func(c *fiber.Ctx) error {
		// Forces the browser to download the APK instead of returning it as text
		return c.Download("./uploads/app/skin_first_app.apk", "Skin_First.apk")
	})

	// ========== ROLE MANAGEMENT ROUTES ==========
	roleHandler := NewRoleAssignmentHandler(firebaseApp)
	staffHandler := api.NewStaffHandler(firebaseApp)
	imageUploadHandler := api.NewImageUploadHandler(firebaseApp)
	appointmentStatusHandler := api.NewAppointmentStatusHandler(firebaseApp)
	reportHandler := api.NewReportHandler(firebaseApp)

	// Admin-only role management routes
	app.Post("/admin/create-staff",authMiddleware.ValidateToken, RequiresRole("admin"), staffHandler.CreateStaffAccount)
	app.Get("/admin/search-staff", staffHandler.SearchStaff)
	app.Get("/admin/search-patients", api.SearchPatients)

	// Admin-only role management routes
	app.Post("/admin/assign-roles", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.AssignRoles)
	app.Get("/admin/user-roles", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.GetUserRoles)
	app.Get("/admin/list-users", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.ListAllUsers)
	app.Delete("/admin/remove-roles", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.RemoveRoles)

	// ========== GLOBAL ASSET ROUTES ==========
	app.Post("/upload/image", authMiddleware.ValidateToken, imageUploadHandler.UploadImage)

	// ========== USER REGISTRATION ROUTES ==========
	// Patient registers themselves after Firebase sign-up (public — no token needed here,
	// but the FirebaseUID in the body ties the record to their auth identity).
	app.Post("/register/patient", api.CreatePatientUser)
	// Only an existing admin can onboard a new doctor or another admin.
	app.Post("/register/doctor-user", api.CreateDoctorUserAccount)
	app.Post("/register/admin", api.CreateAdminUser)

	// ========== ROLE LOOKUP ROUTES ==========
	// Portal: checks admin_users collection only — returns 404 if user is not an admin.
	app.Get("/role/admin", api.FindAdminRole)
	// Mobile app: checks patients + doctor_users collections.
	app.Get("/role/mobile", api.FindMobileUserRole)


	// ========== FOCUS ROUTES ==========
	app.Post("/focus", authMiddleware.ValidateToken, api.CreateFocus)
	app.Get("/findAll/focus", api.GetAllFocuses)
	
	// ========== SERVICE ROUTES ==========
	app.Post("/admin/services", api.CreateService)
	app.Get("/services", api.GetAllServices)
	app.Put("/admin/services/serviceId",api.UpdateService)
	app.Delete("/admin/services/serviceId", api.DeleteService)

	// ========== DOCTOR ROUTES ==========
	app.Post("/doctor", authMiddleware.ValidateToken, RequiresRole("admin"), api.CreateDoctor)
	app.Get("/doctors", authMiddleware.ValidateToken, api.FindAllDoctors)
	app.Get("/doctors/search", api.SearchDoctorInfo)
	app.Get("/findAll/doctors/focus", api.GetDoctorsByFocus)
	app.Get("/doctor-info", authMiddleware.ValidateToken, api.FindDoctorInfoByName)
	app.Get("/doctor-info/id", authMiddleware.ValidateToken, api.FindDoctorInfoByDoctorId)
	app.Put("/doctor-info/id", authMiddleware.ValidateToken, RequiresRole("admin"), api.UpdateDoctorInfoByDoctorId)
	app.Post("/doctor-info", authMiddleware.ValidateToken, RequiresRole("admin"), api.CreateDoctorInfo)

	// Public doctor routes
	app.Get("/doctor-info/favorite", api.GetFavoriteDoctors)
	app.Put("/doctor-info/favorite", api.ToggleFavoriteDoctor)

	// ========== FCM TOKEN ROUTES ==========
	app.Post("/api/users/save-token", api.SaveFCMToken)

	// ========== APPOINTMENT ROUTES ==========
	app.Get("/appointment/next-number/doctorId", api.GetNextAppointmentNumber)
	app.Get("/appointments/running/doctorId", api.GetRunningAppointmentNumber)
	app.Patch("/appointments/id/running", api.SetAppointmentRunning)
	app.Post("/create/appointment", api.CreateAppointment)
	app.Get("/findAll/appointments", authMiddleware.ValidateToken, api.GetAllAppointments)
	app.Get("/findAll/appointments/doctor", authMiddleware.ValidateToken, api.GetAppointmentsByDoctorID)
	app.Get("/findAll/appointments/doctor/ordered", api.GetAppointmentsByDoctorIDSortedByNumber)
	app.Get("/findAll/appointments/doctor/detailed", api.GetAppointmentsByDoctorDateStatus)
	app.Get("/findAll/appointments/patient", api.GetAppointmentsByPatientID)
	app.Get("/appointments/id/appointmentId", api.GetAppointmentByID)
	app.Put("/appointments/id/reschedule", appointmentStatusHandler.RescheduleAppointment)
	app.Patch("/appointments/id/status", appointmentStatusHandler.UpdateAppointmentStatus)

	// ========== DOCTOR SCHEDULE ROUTES ==========
	app.Post("/doctor-schedule", api.CreateDoctorSchedule)
	app.Get("/doctor-schedule", api.GetDoctorSchedule)
	app.Get("/doctor-schedule/range", api.GetDoctorScheduleByDateRange)
	app.Delete("/doctor-schedule", api.DeleteDoctorSchedule)
	app.Delete("/doctor-schedule/time-slot", api.DeleteTimeSlotFromSchedule)

	// ========== MEDICINE ROUTES ==========
	app.Post("/medicines", api.CreateMedicine)
	app.Get("/medicines/search", api.SearchMedicines)
	app.Get("/medicines/low-stock", api.GetLowStockMedicines)
	app.Get("/medicines/:id", api.GetMedicineByID)
	app.Put("/medicines/:id", api.UpdateMedicine)
	app.Delete("/medicines/:id", api.DeleteMedicine)

	// ========== MEDICINE BATCH ROUTES ==========
	app.Post("/batches", api.CreateMedicineBatch)
	app.Get("/batches/medicineId", api.GetBatchesByMedicineID)
	app.Get("/batches/available/medicineId", api.GetAvailableBatchesFEFO)
	app.Get("/batches/active-stock/medicineId", api.GetActiveStockByMedicineID)


	// ========== BILLING ROUTES ==========
	app.Post("/billing/deduct", api.DeductStockFEFO)
	app.Post("/billing/create-bill", api.CreateBill)
	app.Post("/billing/confirm/billId", api.ConfirmBill)
	app.Get("/billing/pdf", api.GenerateBillPDF)

	// Hospital Bill Routes
	app.Post("/billing/hospital-bill", api.CreateHospitalBill)
	app.Put("/billing/hospital-bill/confirm/:id", api.ConfirmHospitalBill)
	app.Get("/billing/hospital-bill/:id/pdf", api.DownloadHospitalBillPDF)

	// ========== MEDICINE ORDER ROUTES ==========
	app.Post("/medicine-orders", api.CreateMedicineOrder)
	app.Get("/medicine-orders/:id", api.GetMedicineOrder)
	app.Get("/medicine-orders", api.SearchMedicineOrders)
	app.Patch("/medicine-orders/:id", api.UpdateMedicineOrderStatus)

	// ========== NEW DOCTOR SCHEDULING ROUTES ==========
	// Doctor Weekly Schedule
	app.Post("/doctor-weekly-schedule", api.CreateDoctorWeeklySchedule)
	app.Put("/doctor-weekly-schedule/doctorId", api.UpdateDoctorWeeklySchedule)
	app.Delete("/doctor-weekly-schedule/doctorId", api.DeleteDoctorWeeklySchedule)
	app.Get("/doctor-weekly-schedule", api.GetAllDoctorWeeklySchedules)
	app.Get("/doctor-weekly-schedule/available-dates", api.GetDoctorAvailableDatesForWeek)

	// Doctor Availability
	app.Post("/doctor-availability", api.CreateDoctorAvailability)
	app.Put("/doctor-availability/doctorAvailabilityId", api.UpdateDoctorAvailability)
	app.Delete("/doctor-availability/doctorAvailabilityId", api.DeleteDoctorAvailability)
	app.Get("/doctor-availability", api.GetAllDoctorAvailabilities)
	app.Get("/doctor-availability/check", api.CheckDoctorAvailability)

	// ========== REPORT ROUTES ==========
	app.Post("/api/reports/upload", reportHandler.UploadReport)
	app.Get("/api/reports", reportHandler.GetReportsByPatientID)

	// ========== NOTIFICATION ROUTES ==========
	// Notifications are created INTERNALLY by the backend — not via a public endpoint.
	// Use functions.SaveAndSendNotification(...) wherever you trigger a notification.
	//
	// GET    /api/notifications?userId=&lastId=&limit= → cursor-based pagination (mobile)
	// PATCH  /api/notifications/:id/read              → mark single as read (mobile)
	// PATCH  /api/notifications/read-all?userId=      → mark all as read (mobile)
	app.Get("/api/notifications", authMiddleware.ValidateToken, api.GetNotifications)
	app.Patch("/api/notifications/:notificationId/read", authMiddleware.ValidateToken, api.MarkNotificationRead)
	app.Patch("/api/notifications/read-all", authMiddleware.ValidateToken, api.MarkAllNotificationsRead)
}
