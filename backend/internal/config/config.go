package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	UploadDir   string
	// S3/MinIO storage
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3PublicURL string
	S3UseSSL    bool
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "5200"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://tattoo:tattoo_dev@localhost:5437/tattoo_consultation?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "tattoo-dev-secret-change-in-prod"),
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
		S3Endpoint:  getEnv("S3_ENDPOINT", "localhost:9010"),
		S3AccessKey: getEnv("S3_ACCESS_KEY", "tattoo-backend"),
		S3SecretKey: getEnv("S3_SECRET_KEY", "tattoo-backend-secret"),
		S3Bucket:    getEnv("S3_BUCKET", "tattoo-consultation"),
		S3PublicURL: getEnv("S3_PUBLIC_URL", "http://localhost:9010"),
		S3UseSSL:    getEnv("S3_USE_SSL", "false") == "true",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
