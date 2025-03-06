package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/nginH/internal/server"
	logs "github.com/nginH/pkg/log"
)

func main() {
	logs.InitLogger()

	port := os.Getenv("PORT")
	if port == "" {
		logs.Error("PORT is not set")
		port = "6969"
		logs.Info("Setting PORT to default value: ", port)
	}

	// Create and start server
	srv := server.New()
	if err := srv.Start(); err != nil {
		logs.Fatal("Failed to start server:", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	logs.Info("Shutting down server...")
	if err := srv.Stop(); err != nil {
		logs.Error("Error during shutdown:", err)
	}
}
