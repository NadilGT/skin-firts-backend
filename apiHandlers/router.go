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

	// ========== ROLE MANAGEMENT ROUTES ==========
	roleHandler := NewRoleAssignmentHandler(firebaseApp)
	imageUploadHandler := api.NewImageUploadHandler(firebaseApp)

	// 🚨 Call once to create first admin: /admin/initialize?email=you@example.com
	app.Get("/admin/initialize", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.InitializeSuperAdmin)

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

	// ========== DOCTOR ROUTES ==========
	app.Post("/doctor", authMiddleware.ValidateToken, RequiresRole("admin"), api.CreateDoctor)
	app.Get("/doctors", authMiddleware.ValidateToken, api.FindAllDoctors)
	app.Get("/findAll/doctors/focus", api.GetDoctorsByFocus)
	app.Get("/doctor-info", authMiddleware.ValidateToken, api.FindDoctorInfoByName)
	app.Post("/doctor-info", authMiddleware.ValidateToken, RequiresRole("admin"), api.CreateDoctorInfo)

	// Public doctor routes
	app.Get("/doctor-info/favorite", api.GetFavoriteDoctors)
	app.Put("/doctor-info/favorite", api.ToggleFavoriteDoctor)

	// ========== APPOINTMENT ROUTES ==========
	app.Get("/appointment/next-number/doctorId", api.GetNextAppointmentNumber)
	app.Get("/appointments/running/doctorId", api.GetRunningAppointmentNumber)
	app.Patch("/appointments/id/running", api.SetAppointmentRunning)
	app.Post("/create/appointment", api.CreateAppointment)
	app.Get("/findAll/appointments", authMiddleware.ValidateToken, api.GetAllAppointments)
	app.Get("/findAll/appointments/doctor", authMiddleware.ValidateToken, api.GetAppointmentsByDoctorID)
	app.Get("/appointments/id/appointmentId", api.GetAppointmentByID)
	app.Put("/appointments/id/reschedule", api.RescheduleAppointment)
	app.Patch("/appointments/id/status", api.UpdateAppointmentStatus)

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

	// ========== BILLING ROUTES ==========
	app.Post("/billing/deduct", api.DeductStockFEFO)

	// ========== MEDICINE ORDER ROUTES ==========
	app.Post("/medicine-orders", api.CreateMedicineOrder)
	app.Get("/medicine-orders/:id", api.GetMedicineOrder)
	app.Get("/medicine-orders", api.SearchMedicineOrders)
	app.Patch("/medicine-orders/:id", api.UpdateMedicineOrderStatus)
}
