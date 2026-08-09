package config

import "os"

const (
	defaultPort    = "8080"
	defaultGinMode = "debug"
)

type Config struct {
	Port        string
	GinMode     string
	DatabaseURL string
}

func Load() Config {
	return Config{
		Port:        getEnv("APP_PORT", defaultPort),
		GinMode:     getEnv("GIN_MODE", defaultGinMode),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}
