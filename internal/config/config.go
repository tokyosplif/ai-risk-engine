package config

import (
	"os"
)

type Config struct {
	Port        string
	Groq        GroqConfig
	PromptsPath string
}

type GroqConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", ":50051"),
		PromptsPath: getEnv("PROMPTS_PATH", "prompts.json"),
		Groq: GroqConfig{
			APIKey:  os.Getenv("GROQ_API_KEY"),
			BaseURL: getEnv("GROQ_BASE_URL", "https://api.groq.com/openai/v1"),
			Model:   getEnv("GROQ_MODEL", "llama-3.3-70b-versatile"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
