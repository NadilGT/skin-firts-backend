package main

import (
	"lawyerSL-Backend/apiHandlers"
	"lawyerSL-Backend/dbConfigs"
	"log"

	"github.com/gofiber/fiber/v2"
)

func main(){
	app := fiber.New()

	dbConfigs.ConnectMongoDB("mongodb+srv://admin:W6ptbj7HPS3RJ4cU@cluster0.tgypip5.mongodb.net/")

	apiHandlers.SetupRoutes(app)

	log.Fatal(app.Listen(":3000"))
}