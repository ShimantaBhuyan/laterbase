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
	Crawl4AIURL string
}

func Load() *Config {
	// Best-effort load .env — ignore if not present
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL: envOrDefault("DATABASE_URL", "postgres://linkgroupy:linkgroupy_dev@localhost:5433/linkgroupy?sslmode=disable"),
		Port:        envOrDefault("PORT", "8080"),
		LLMAPIKey:   os.Getenv("LLM_API_KEY"),
		LLMModel:    envOrDefault("LLM_MODEL", "gpt-4o-mini"),
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
