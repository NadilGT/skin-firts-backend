package apiHandlers

import (
	"lawyerSL-Backend/api"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, authMiddleware *AuthMiddleware) {

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello from Fiber on Render!")
	})

	app.Post("/doctor", authMiddleware.ValidateToken, api.CreateDoctor)
	app.Get("/doctors", authMiddleware.ValidateToken, api.FindAllDoctors)
	app.Get("/doctor-info", api.FindDoctorInfoByName)
	app.Post("/doctor-info", api.CreateDoctorInfo)
}
