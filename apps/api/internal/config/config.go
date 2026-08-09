package config

import "os"

const defaultPort = "8080"

type Config struct {
	Port    string
	GinMode string
}

func Load() Config {
	return Config{
		Port:    getEnv("APP_PORT", defaultPort),
		GinMode: getEnv("GIN_MODE", "debug"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}
