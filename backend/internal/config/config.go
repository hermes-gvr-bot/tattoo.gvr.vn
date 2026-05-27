package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	UploadDir   string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "5200"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://tattoo:tattoo_dev@localhost:5437/tattoo_consultation?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "tattoo-dev-secret-change-in-prod"),
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
