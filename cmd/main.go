package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/nginH/internal/server"
	_ "github.com/nginH/internal/server/cache_middleware"
	logs "github.com/nginH/pkg/log"
)

func main() {
	logs.InitLogger()
	if err := godotenv.Load(); err != nil {
		logs.Warn("No .env file found, using environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "6969"
		logs.Info("PORT not set, using default:", port)
	}

	srv := server.New()
	if err := srv.Start(); err != nil {
		logs.Fatal("Failed to start server:", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	logs.Info("Shutting down server...")
	if err := srv.Stop(); err != nil {
		logs.Error("Error during shutdown:", err)
	}
}
