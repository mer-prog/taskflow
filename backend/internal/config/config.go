package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port             string
	Env              string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	DBURL_           string
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration
	CORSOrigins      []string
}

func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "8080"),
		Env:              getEnv("ENV", "development"),
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "taskflow"),
		DBPassword:       getEnv("DB_PASSWORD", "taskflow"),
		DBName:           getEnv("DB_NAME", "taskflow"),
		DBSSLMode:        getEnv("DB_SSLMODE", "disable"),
		DBURL_:           getEnv("DATABASE_URL", ""),
		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTAccessExpiry:  parseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m")),
		JWTRefreshExpiry: parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h")),
		CORSOrigins:      parseCORSOrigins(getEnv("CORS_ORIGIN", "http://localhost:3000")),
	}
}

func (c *Config) DBURL() string {
	if c.DBURL_ != "" {
		return c.DBURL_
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func parseCORSOrigins(s string) []string {
	parts := strings.Split(s, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}
