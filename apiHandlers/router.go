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

	// 🚨 Call once to create first admin: /admin/initialize?email=you@example.com
	app.Get("/admin/initialize", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.InitializeSuperAdmin)

	// Admin-only role management routes
	app.Post("/admin/assign-roles", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.AssignRoles)
	app.Get("/admin/user-roles", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.GetUserRoles)
	app.Get("/admin/list-users", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.ListAllUsers)
	app.Delete("/admin/remove-roles", authMiddleware.ValidateToken, RequiresRole("admin"), roleHandler.RemoveRoles)

	// ========== DOCTOR ROUTES ==========
	app.Post("/doctor", authMiddleware.ValidateToken, RequiresRole("admin"), api.CreateDoctor)
	app.Get("/doctors", authMiddleware.ValidateToken, api.FindAllDoctors)
	app.Get("/doctor-info", authMiddleware.ValidateToken, api.FindDoctorInfoByName)
	app.Post("/doctor-info", authMiddleware.ValidateToken, RequiresRole("admin"), api.CreateDoctorInfo)

	// Public doctor routes
	app.Get("/doctor-info/favorite", api.GetFavoriteDoctors)
	app.Put("/doctor-info/favorite", api.ToggleFavoriteDoctor)

	// ========== APPOINTMENT ROUTES ==========
	app.Post("/create/appointment", api.CreateAppointment)
	app.Get("/findAll/appointments", authMiddleware.ValidateToken, api.GetAllAppointments)
	app.Put("/appointments/:id/reschedule", api.RescheduleAppointment)
	app.Patch("/appointments/:id/status", api.UpdateAppointmentStatus)

	// ========== DOCTOR SCHEDULE ROUTES ==========
	app.Post("/doctor-schedule", api.CreateDoctorSchedule)
	app.Get("/doctor-schedule", api.GetDoctorSchedule)
	app.Get("/doctor-schedule/range", api.GetDoctorScheduleByDateRange)
	app.Delete("/doctor-schedule", api.DeleteDoctorSchedule)
	app.Delete("/doctor-schedule/time-slot", api.DeleteTimeSlotFromSchedule)

	// ========== MEDICINE ROUTES ==========
	app.Post("/medicines",authMiddleware.ValidateToken, api.CreateMedicine)
	app.Get("/medicines/search", api.SearchMedicines)
	app.Get("/medicines/low-stock", api.GetLowStockMedicines)
	app.Get("/medicines/:id", api.GetMedicineByID)
	app.Put("/medicines/:id", api.UpdateMedicine)
	app.Delete("/medicines/:id", api.DeleteMedicine)

	// ========== MEDICINE ORDER ROUTES ==========
	app.Post("/medicine-orders", api.CreateMedicineOrder)
	app.Get("/medicine-orders/:id", api.GetMedicineOrder)
	app.Get("/medicine-orders", api.SearchMedicineOrders)
	app.Patch("/medicine-orders/:id", api.UpdateMedicineOrderStatus)
}
