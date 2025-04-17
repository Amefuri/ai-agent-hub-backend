package main

import (
	"fmt"
	"log"
	"math/rand"

	"ai-agent-hub/internal/database"
	"ai-agent-hub/internal/models"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/joho/godotenv"
)

func main() {

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	
	// Connect to DB
	db := database.Connect()

	database.Migrate(db)

	// Wipe existing data (for testing only)
	db.Exec("DELETE FROM agents")
	db.Exec("DELETE FROM users")

	for i := 1; i <= 5; i++ {
		user := models.User{
			Username: fmt.Sprintf("user%d", i),
			Email:    fmt.Sprintf("user%d@example.com", i),
			Password: "hashed-password", // Replace with actual hash if needed
		}

		for j := 1; j <= 5; j++ {
			agent := models.Agent{
				Name:          fmt.Sprintf("Agent %d-%d", i, j),
				Description:   gofakeit.HipsterSentence(10),
				Avatar:        fmt.Sprintf("https://api.dicebear.com/7.x/identicon/svg?seed=%d%d", i, j),
				SystemPrompt:  gofakeit.Sentence(5),
				InputTemplate: "{{input}}",
				Personality:   []string{"friendly", "serious", "sarcastic", "formal"}[rand.Intn(4)],
			}
			user.Agents = append(user.Agents, agent)
		}

		err := db.Create(&user).Error
		if err != nil {
			panic(err)
		}
		fmt.Printf("✅ Created %s with %d agents\n", user.Username, len(user.Agents))
	}
}
