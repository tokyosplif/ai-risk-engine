package main

import (
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/tokyosplif/ai-risk-engine/internal/app"
	"github.com/tokyosplif/ai-risk-engine/pkg/logger"
)

func main() {
	logger.Setup()

	if err := godotenv.Load(); err != nil {
		slog.Info(".env not found, using system environment variables")
	}

	if err := app.RunServer(); err != nil {
		slog.Error("Server failed to start", "error", err)
	}
}
