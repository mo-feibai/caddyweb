package main

import (
	"fmt"
	"log"
	"os"

	"github.com/caddyweb/caddyapi/internal/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	router := handlers.NewRouter()

	fmt.Printf("🚀 CaddyAPI server starting on port %s\n", port)
	fmt.Println("📡 API endpoints available at http://localhost:" + port + "/api")

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
