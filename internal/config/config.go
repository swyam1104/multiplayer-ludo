package config

import (
	"os"
)

type Config struct {
	Port      string
	JWTSecret string
	DBDSN     string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecretjwtkey"
	}

	return &Config{
		Port:      port,
		JWTSecret: jwtSecret,
		DBDSN:     os.Getenv("DB_DSN"),
	}
}
