package apiHandlers

import (
	"lawyerSL-Backend/api"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello from Fiber on Render!")
	})

	app.Post("/doctor", api.CreateDoctor)
	app.Get("/doctors", api.FindAllDoctors)
}
