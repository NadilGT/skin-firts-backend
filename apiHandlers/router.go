package apiHandlers

import (
	"lawyerSL-Backend/api"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, authMiddleware *AuthMiddleware) {

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello from Fiber on Render!")
	})

	app.Post("/doctor", api.CreateDoctor)
	app.Get("/doctors", api.FindAllDoctors)
	app.Get("/doctor-info", api.FindDoctorInfoByName)
	app.Post("/doctor-info", api.CreateDoctorInfo)
	app.Get("/doctor-info/favorite", api.GetFavoriteDoctors)
	app.Put("/doctor-info/favorite", api.ToggleFavoriteDoctor)
	app.Post("/create/appointment", api.CreateAppointment)
	app.Get("/findAll/appointments", api.GetAllAppointments)
	app.Put("/appointments/:id/reschedule", api.RescheduleAppointment)
	app.Patch("/appointments/:id/status", api.UpdateAppointmentStatus)

	app.Post("/doctor-schedule", api.CreateDoctorSchedule)
	app.Get("/doctor-schedule", api.GetDoctorSchedule)
	app.Get("/doctor-schedule/range", api.GetDoctorScheduleByDateRange)
	app.Delete("/doctor-schedule", api.DeleteDoctorSchedule)
	app.Delete("/doctor-schedule/time-slot", api.DeleteTimeSlotFromSchedule)


	app.Post("/medicines", api.CreateMedicine)
	app.Get("/medicines/search", api.SearchMedicines)
	app.Get("/medicines/low-stock", api.GetLowStockMedicines)
	app.Get("/medicines/:id", api.GetMedicineByID)
	app.Put("/medicines/:id", api.UpdateMedicine)
	app.Delete("/medicines/:id", api.DeleteMedicine)

	app.Post("/medicine-orders", api.CreateMedicineOrder)
	app.Get("/medicine-orders/:id", api.GetMedicineOrder)
	app.Get("/medicine-orders", api.SearchMedicineOrders)
	app.Patch("/medicine-orders/:id", api.UpdateMedicineOrderStatus)
}
