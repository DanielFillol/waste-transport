package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port               string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPass             string
	DBName             string
	JWTSecret          string
	JWTExpirationHours int
	OpenAIAPIKey       string
	OpenAIModel        string
	AppEnv             string
}

func Load() *Config {
	jwtExp, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "8"))
	return &Config{
		Port:               getEnv("PORT", "8080"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPass:             getEnv("DB_PASS", "postgres"),
		DBName:             getEnv("DB_NAME", "waste_db"),
		JWTSecret:          getEnv("JWT_SECRET", "change-me"),
		JWTExpirationHours: jwtExp,
		OpenAIAPIKey:       getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:        getEnv("OPENAI_MODEL", "gpt-4o"),
		AppEnv:             getEnv("APP_ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
