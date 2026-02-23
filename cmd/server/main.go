package main

import (
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/tokyosplif/ai-risk-engine/internal/app"
	"github.com/tokyosplif/ai-risk-engine/internal/config"
	"github.com/tokyosplif/ai-risk-engine/pkg/logger"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	logger.Setup(cfg.LogLevel)

	if cfg.Groq.APIKey == "" {
		slog.Error("CRITICAL: GROQ_API_KEY is missing")
		return
	}

	slog.Info("Starting AI Risk Engine", "env", "prod", "model", cfg.Groq.Model)

	if err := app.RunServer(cfg); err != nil {
		slog.Error("Application failed", "error", err)
	}
}
