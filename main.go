package main

import (
	"lawyerSL-Backend/apiHandlers"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/dto"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:8080",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE",
	}))

	dbConfigs.ConnectMongoDB("mongodb+srv://admin:W6ptbj7HPS3RJ4cU@cluster0.tgypip5.mongodb.net/")

	firebaseApp, err := apiHandlers.InitFirebaseApp()
	if err != nil {
		log.Fatalf("❌ Failed to initialize Firebase: %v", err)
	}

	authConfig := dto.AuthConfig{
		FirebaseProjectID: os.Getenv("FIREBASE_PROJECT_ID"),
	}

	authMiddleware := apiHandlers.NewAuthMiddleware(authConfig, firebaseApp)

	apiHandlers.SetupRoutes(app, authMiddleware)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(app.Listen("0.0.0.0:" + port))
}