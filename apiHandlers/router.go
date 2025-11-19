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

	app.Post("/doctor-schedule", api.CreateDoctorSchedule)
	app.Get("/doctor-schedule", api.GetDoctorSchedule)
	app.Get("/doctor-schedule/range", api.GetDoctorScheduleByDateRange)
	app.Delete("/doctor-schedule", api.DeleteDoctorSchedule)
}
