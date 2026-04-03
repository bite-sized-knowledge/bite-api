package config

import (
	"log"
	"os"
	"time"
)

type Config struct {
	Port               string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	JWTSecretKey       string
	JWTAccessExpiry    time.Duration
	JWTRefreshExpiry   time.Duration
	ResendAPIKey       string
	EmailFrom          string
	RecsysBaseURL      string
	RecsysAPIKey       string
	AppBaseURL         string
	AppEnv             string
	GitHubClientID     string
	GitHubClientSecret string
	GoogleClientID     string
	GoogleClientSecret string
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		DBHost:             getEnv("DB_HOST", "127.0.0.1"),
		DBPort:             getEnv("DB_PORT", "3306"),
		DBUser:             getEnv("DB_USER", "bite"),
		DBPassword:         getEnv("DB_PASSWORD", "bite"),
		DBName:             getEnv("DB_NAME", "bite"),
		JWTSecretKey:       requireEnvInProd("JWT_SECRET_KEY", "dev-secret"),
		JWTAccessExpiry:    getDurationEnv("JWT_ACCESS_EXPIRY", 15*time.Minute),
		JWTRefreshExpiry:   getDurationEnv("JWT_REFRESH_EXPIRY", 365*24*time.Hour),
		ResendAPIKey:       getEnv("RESEND_API_KEY", ""),
		EmailFrom:          getEnv("EMAIL_FROM", "Bite <noreply@bite-sized.xyz>"),
		RecsysBaseURL:      getEnv("RECSYS_BASE_URL", "http://localhost:8001"),
		RecsysAPIKey:       getEnv("RECSYS_API_KEY", ""),
		AppBaseURL:         getEnv("APP_BASE_URL", "http://localhost:8080"),
		AppEnv:             getEnv("APP_ENV", "local"),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func requireEnvInProd(key, fallback string) string {
	v := os.Getenv(key)
	if v != "" {
		return v
	}
	if env := os.Getenv("APP_ENV"); env == "production" {
		log.Fatalf("FATAL: %s must be set in production", key)
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
