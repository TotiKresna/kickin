package config

import (
	"log"
	"os"
)

type Config struct {
	AppPort     	string
	DBUser      	string
	DBPassword  	string
	DBHost      	string
	DBPort      	string
	DBName      	string
	AllowedOrigins	string
}

func LoadConfig() *Config {
	return &Config{
		AppPort:     getEnvWithDefault("APP_PORT", "8080"),
		DBUser:      getEnv("DB_USER"),
		DBPassword:  getEnv("DB_PASSWORD"),
		DBHost:      getEnv("DB_HOST"),
		DBPort:      getEnv("DB_PORT"),
		DBName:      getEnv("DB_NAME"),
		AllowedOrigins: getEnvWithDefault("ALLOWED_ORIGINS", "http://localhost:3000"),
	}
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Environment variable %s is not set", key)
	}
	return val
}

func getEnvWithDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
