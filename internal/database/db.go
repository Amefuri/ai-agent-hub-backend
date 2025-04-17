package database

import (
	"ai-agent-hub/internal/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() *gorm.DB {

    var dsn string

    // Use DATABASE_URL if available (Render and other cloud providers usually set this)
    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL != "" {
        dsn = databaseURL
    } else {
        // Fallback to local environment variables
        host := os.Getenv("DB_HOST")
        user := os.Getenv("DB_USER")
        password := os.Getenv("DB_PASSWORD")
        dbname := os.Getenv("DB_NAME")
        port := os.Getenv("DB_PORT")

        dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
            host, user, password, dbname, port)
    }

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect to database")
    }
    fmt.Println("✅ Database connected")
    Migrate(db)
    return db
}
func Migrate(db *gorm.DB) {
    err := db.AutoMigrate(&models.User{}, &models.Agent{})
    if err != nil {
        log.Fatal("❌ Failed DB migration:", err)
    }
    fmt.Println("✅ Database migrated")
}