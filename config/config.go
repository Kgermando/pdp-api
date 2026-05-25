package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort         string
	AppEnv          string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	JWTSecret       string
	JWTExpires      string
	UploadDir       string
	MaxUpload       int64
	SMSAPIKey       string
	SMSUser         string
	SMSSender       string
	// Backblaze B2
	B2KeyID         string
	B2ApplicationKey string
	B2BucketID      string
	B2BucketName    string
	// Super Admin seed
	SuperAdminEmail    string
	SuperAdminPassword string
	SuperAdminName     string
}

var AppConfig Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	maxUpload, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE", "10485760"), 10, 64)

	AppConfig = Config{
		AppPort:    getEnv("APP_PORT", "3000"),
		AppEnv:     getEnv("APP_ENV", "production"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "pdp_db"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),
		JWTSecret:  getEnv("JWT_SECRET", "default_secret_change_me"),
		JWTExpires: getEnv("JWT_EXPIRES_IN", "24h"),
		UploadDir:  getEnv("UPLOAD_DIR", "./uploads"),
		MaxUpload:  maxUpload,
		SMSAPIKey:        getEnv("SMS_API_KEY", ""),
		SMSUser:          getEnv("SMS_USERNAME", "sandbox"),
		SMSSender:        getEnv("SMS_SENDER_ID", "PDP"),
		// Backblaze B2
		B2KeyID:          getEnv("BACKBLAZE_KEY_ID", ""),
		B2ApplicationKey: getEnv("BACKBLAZE_APPLICATION_KEY", ""),
		B2BucketID:       getEnv("BACKBLAZE_BUCKET_ID", ""),
		B2BucketName:     getEnv("BACKBLAZE_BUCKET_NAME", ""),
		// Super Admin seed
		SuperAdminEmail:    getEnv("SUPER_ADMIN_EMAIL", "superadmin@pdp.cd"),
		SuperAdminPassword: getEnv("SUPER_ADMIN_PASSWORD", ""),
		SuperAdminName:     getEnv("SUPER_ADMIN_NAME", "Super Administrateur"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
