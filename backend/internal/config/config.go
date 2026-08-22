package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	LLMAPIKey   string
	LLMModel    string
	AskLLMModel string
	Crawl4AIURL string
}

func Load() *Config {
	// Best-effort load .env — ignore if not present
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL: envOrDefault("DATABASE_URL", "postgres://laterbase:laterbase_dev@localhost:5433/laterbase?sslmode=disable"),
		Port:        envOrDefault("PORT", "8080"),
		LLMAPIKey:   os.Getenv("LLM_API_KEY"),
		LLMModel:    envOrDefault("LLM_MODEL", "gpt-4o-mini"),
		AskLLMModel: envOrDefault("ASK_LLM_MODEL", "gpt-5.6-luna"),
		Crawl4AIURL: os.Getenv("CRAWL4AI_URL"),
	}

	return cfg
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
