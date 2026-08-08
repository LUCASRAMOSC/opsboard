package main

import (
	"log"

	"github.com/LUCASRAMOSC/opsboard/apps/api/internal/server"
)

func main() {
	router := server.New()

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
