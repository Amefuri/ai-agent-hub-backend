package main

import (
	"ai-agent-hub/internal/database"
	"ai-agent-hub/internal/routes"
	"ai-agent-hub/internal/utils"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func loadEnv() {
    // Only load .env in development, not in production
    if os.Getenv("ENV") != "production" {
        err := godotenv.Load()
        if err != nil {
            log.Fatal("Error loading .env file")
        }
    }
}

func main() {

    loadEnv()
    
    db := database.Connect()
    e := echo.New()
    e.Validator = utils.NewValidator()

    // Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

    routes.RegisterPublicRoutes(e, db)
    routes.RegisterPrivateRoutes(e, db)

    port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

    e.Logger.Fatal(e.Start(":" + port))
}