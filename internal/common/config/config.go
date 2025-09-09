package config

import (
	"os"
	"log"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string

	// MinIO
	MinIOEndpoint   string
	MinIOAccessKey  string
	MinIOSecretKey  string
	MinIOUseSSL     bool

	// JWT
	JWTSecret string

	// Payment Gateway
	RazorpayKeyID     string
	RazorpayKeySecret string

	// Server
	ServerPort string

	// Logging
	GinMode  string
	LogLevel string
}

var AppConfig *Config

func LoadConfig() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Parse boolean values
	minioUseSSL, _ := strconv.ParseBool(getEnv("MINIO_USE_SSL", "false"))

	AppConfig = &Config{
		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3336"),
		DBName:     getEnv("DB_NAME", "gosocial_db"),
		DBUser:     getEnv("DB_USER", "gosocial_user"),
		DBPassword: getEnv("DB_PASSWORD", "G0Social@123"),

		// Redis
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", "redis_pass_2024"),

		// MinIO
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin123"),
		MinIOUseSSL:    minioUseSSL,

		// JWT
		JWTSecret: getEnv("JWT_SECRET", "commerce_jwt_secret_2024"),

		// Payment Gateway
		RazorpayKeyID:     getEnv("RAZORPAY_KEY_ID", ""),
		RazorpayKeySecret: getEnv("RAZORPAY_KEY_SECRET", ""),

		// Server
		ServerPort: getEnv("SERVER_PORT", "8080"),

		// Logging
		GinMode:  getEnv("GIN_MODE", "debug"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	log.Printf("Configuration loaded successfully")
	log.Printf("Database: %s:%s/%s", AppConfig.DBHost, AppConfig.DBPort, AppConfig.DBName)
	log.Printf("Redis: %s:%s", AppConfig.RedisHost, AppConfig.RedisPort)
	log.Printf("MinIO: %s", AppConfig.MinIOEndpoint)
	log.Printf("Server will run on port: %s", AppConfig.ServerPort)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper functions for specific configs
func GetDatabaseDSN() string {
	return AppConfig.DBUser + ":" + AppConfig.DBPassword + 
	"@tcp(" + AppConfig.DBHost + ":" + AppConfig.DBPort + ")/" + 
	AppConfig.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
}

func GetRedisAddress() string {
	return AppConfig.RedisHost + ":" + AppConfig.RedisPort
}

func IsProduction() bool {
	return getEnv("GIN_MODE", "debug") == "release"
}

