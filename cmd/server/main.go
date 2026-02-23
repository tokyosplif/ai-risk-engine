package main

import (
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/tokyosplif/ai-risk-engine/internal/app"
	"github.com/tokyosplif/ai-risk-engine/internal/config"
	"github.com/tokyosplif/ai-risk-engine/pkg/logger"
)

func main() {
	logger.Setup()

	if err := godotenv.Load(); err != nil {
		slog.Debug(".env not found, using system environment variables")
	}

	cfg := config.Load()

	if cfg.Groq.APIKey == "" {
		slog.Error("CRITICAL: GROQ_API_KEY is empty. AI features will not work.")
		return
	}

	if err := app.RunServer(cfg); err != nil {
		slog.Error("Server failed to start", "error", err)
	}
}
