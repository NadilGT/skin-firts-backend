package main

import (
	localauth "lawyerSL-Backend/auth"
	"lawyerSL-Backend/apiHandlers"
	"lawyerSL-Backend/dao"
	"lawyerSL-Backend/dbConfigs"
	"lawyerSL-Backend/integrations"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	app := fiber.New()

	_ = godotenv.Load(".env")
	integrations.SetEnvironmentVariables()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000,http://localhost:5173,http://localhost:8080,https://med-center-hub.vercel.app,https://medical.nadildinsara.me,https://*.vercel.app,https://medical-portal.codekongsl.com",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, PATCH",
	}))

	// Serve built frontend (React/Vite) from ./frontend/dist
	app.Static("/", "./frontend/dist")

	// Serve uploaded files (images, reports) — used in offline/local mode
	app.Static("/uploads", "./uploads")

	// Connect to MongoDB
	dbConfigs.ConnectMongoDB(os.Getenv("MONGODB_URI"))

	// Start background jobs
	dao.StartBillExpiryCron()

	// Seed super admin on startup (local auth — no Firebase)
	localauth.InitializeSuperAdmin()

	// Build JWT auth middleware
	authMiddleware := apiHandlers.NewAuthMiddleware()

	// Register all routes
	apiHandlers.SetupRoutes(app, authMiddleware)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("🚀 Server starting on port %s (Local Auth Mode — No Firebase)\n", port)
	log.Fatal(app.Listen("0.0.0.0:" + port))
}
