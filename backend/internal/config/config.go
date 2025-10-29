package config

import (
	"os"
	"strings"
)

type Config struct {
	Port         string
	DatabaseURL  string
	KafkaBrokers []string
	KafkaEnabled bool
}

func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/four_in_a_row?sslmode=disable"),
		KafkaBrokers: []string{getEnv("KAFKA_BROKER", "localhost:9092")},
		KafkaEnabled: strings.ToLower(getEnv("KAFKA_ENABLED", "true")) == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
