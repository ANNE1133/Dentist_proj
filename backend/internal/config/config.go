package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	PythonServiceURL string
}

func Load() Config {
	return Config{
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "dentist"),
		DBPassword:       getEnv("DB_PASSWORD", "dentistpass"),
		DBName:           getEnv("DB_NAME", "dentistchair"),
		PythonServiceURL: getEnv("PYTHON_SERVICE_URL", "http://host.docker.internal:8000"),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}