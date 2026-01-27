package main

import (
	"log"

	"github.com/michelemendel/times.place/internal/http"
)

func main() {
	// Create and start server
	server, err := http.NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server (blocks until shutdown)
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
