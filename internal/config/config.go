package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	MinIO      MinIOConfig
	JWT        JWTConfig
	OTP        OTPConfig
	TwoFA      TwoFAConfig
	SMTP       SMTPConfig
	CORS       CORSConfig
	RateLimit  RateLimitConfig
	FileUpload FileUploadConfig
	Logging    LoggingConfig
	Pagination PaginationConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

// MinIOConfig holds MinIO configuration
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret             string
	AccessTokenExpire  time.Duration
	RefreshTokenExpire time.Duration
}

// OTPConfig holds OTP configuration
type OTPConfig struct {
	Expire time.Duration
	Length int
}

// TwoFAConfig holds 2FA configuration
type TwoFAConfig struct {
	Issuer string
}

// SMTPConfig holds SMTP email configuration
type SMTPConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	FromName  string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Requests int
	Duration time.Duration
}

// FileUploadConfig holds file upload configuration
type FileUploadConfig struct {
	MaxUploadSize    int64
	AllowedFileTypes []string
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
	Output string
}

// PaginationConfig holds pagination configuration
type PaginationConfig struct {
	DefaultPageSize int
	MaxPageSize     int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Get environment (default to development)
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	// Try to load environment-specific config file
	envFile := fmt.Sprintf(".env.%s", env)
	if err := viper.ReadInConfig(); err == nil {
		// .env file exists, use it
		viper.SetConfigFile(".env")
	} else {
		// Try environment-specific file
		viper.SetConfigFile(envFile)
	}

	// Read config file (don't fail if it doesn't exist)
	if err := viper.ReadInConfig(); err != nil {
		// If no config file, that's okay - we'll use env vars and defaults
		fmt.Printf("No config file found, using environment variables and defaults\n")
	} else {
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}

	// Always read from environment variables (they override config file)
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	config := &Config{
		Server: ServerConfig{
			Host:         viper.GetString("SERVER_HOST"),
			Port:         viper.GetInt("SERVER_PORT"),
			ReadTimeout:  viper.GetDuration("SERVER_READ_TIMEOUT"),
			WriteTimeout: viper.GetDuration("SERVER_WRITE_TIMEOUT"),
		},
		Database: DatabaseConfig{
			Host:            viper.GetString("DB_HOST"),
			Port:            viper.GetInt("DB_PORT"),
			User:            viper.GetString("DB_USER"),
			Password:        viper.GetString("DB_PASSWORD"),
			DBName:          viper.GetString("DB_NAME"),
			SSLMode:         viper.GetString("DB_SSLMODE"),
			MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: viper.GetDuration("DB_CONN_MAX_LIFETIME"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetInt("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
			PoolSize: viper.GetInt("REDIS_POOL_SIZE"),
		},
		MinIO: MinIOConfig{
			Endpoint:  viper.GetString("MINIO_ENDPOINT"),
			AccessKey: viper.GetString("MINIO_ACCESS_KEY"),
			SecretKey: viper.GetString("MINIO_SECRET_KEY"),
			UseSSL:    viper.GetBool("MINIO_USE_SSL"),
			Bucket:    viper.GetString("MINIO_BUCKET"),
		},
		JWT: JWTConfig{
			Secret:             viper.GetString("JWT_SECRET"),
			AccessTokenExpire:  viper.GetDuration("JWT_ACCESS_TOKEN_EXPIRE"),
			RefreshTokenExpire: viper.GetDuration("JWT_REFRESH_TOKEN_EXPIRE"),
		},
		OTP: OTPConfig{
			Expire: viper.GetDuration("OTP_EXPIRE"),
			Length: viper.GetInt("OTP_LENGTH"),
		},
		TwoFA: TwoFAConfig{
			Issuer: viper.GetString("TWO_FA_ISSUER"),
		},
		SMTP: SMTPConfig{
			Host:      viper.GetString("SMTP_HOST"),
			Port:      viper.GetInt("SMTP_PORT"),
			Username:  viper.GetString("SMTP_USERNAME"),
			Password:  viper.GetString("SMTP_PASSWORD"),
			FromEmail: viper.GetString("SMTP_FROM_EMAIL"),
			FromName:  viper.GetString("SMTP_FROM_NAME"),
		},
		CORS: CORSConfig{
			AllowedOrigins:   viper.GetStringSlice("CORS_ALLOWED_ORIGINS"),
			AllowedMethods:   viper.GetStringSlice("CORS_ALLOWED_METHODS"),
			AllowedHeaders:   viper.GetStringSlice("CORS_ALLOWED_HEADERS"),
			AllowCredentials: viper.GetBool("CORS_ALLOW_CREDENTIALS"),
		},
		RateLimit: RateLimitConfig{
			Requests: viper.GetInt("RATE_LIMIT_REQUESTS"),
			Duration: viper.GetDuration("RATE_LIMIT_DURATION"),
		},
		FileUpload: FileUploadConfig{
			MaxUploadSize:    viper.GetInt64("MAX_UPLOAD_SIZE"),
			AllowedFileTypes: viper.GetStringSlice("ALLOWED_FILE_TYPES"),
		},
		Logging: LoggingConfig{
			Level:  viper.GetString("LOG_LEVEL"),
			Format: viper.GetString("LOG_FORMAT"),
			Output: viper.GetString("LOG_OUTPUT"),
		},
		Pagination: PaginationConfig{
			DefaultPageSize: viper.GetInt("DEFAULT_PAGE_SIZE"),
			MaxPageSize:     viper.GetInt("MAX_PAGE_SIZE"),
		},
	}

	return config, nil
}

// setDefaults sets default values for configuration
func setDefaults() {
	// Server defaults
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("SERVER_PORT", 8080)
	viper.SetDefault("SERVER_READ_TIMEOUT", "30s")
	viper.SetDefault("SERVER_WRITE_TIMEOUT", "30s")

	// Database defaults
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "go_cms")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 5)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", "5m")

	// Redis defaults
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("REDIS_POOL_SIZE", 10)

	// MinIO defaults
	viper.SetDefault("MINIO_ENDPOINT", "localhost:9000")
	viper.SetDefault("MINIO_ACCESS_KEY", "minioadmin")
	viper.SetDefault("MINIO_SECRET_KEY", "minioadmin")
	viper.SetDefault("MINIO_USE_SSL", false)
	viper.SetDefault("MINIO_BUCKET", "go-cms")

	// JWT defaults
	viper.SetDefault("JWT_SECRET", "change-this-secret-key")
	viper.SetDefault("JWT_ACCESS_TOKEN_EXPIRE", "15m")
	viper.SetDefault("JWT_REFRESH_TOKEN_EXPIRE", "168h") // 7 days

	// OTP defaults
	viper.SetDefault("OTP_EXPIRE", "30s")
	viper.SetDefault("OTP_LENGTH", 6)

	// 2FA defaults
	viper.SetDefault("TWO_FA_ISSUER", "GO-CMS")

	// SMTP defaults
	viper.SetDefault("SMTP_HOST", "smtp.gmail.com")
	viper.SetDefault("SMTP_PORT", 587)
	viper.SetDefault("SMTP_USERNAME", "")
	viper.SetDefault("SMTP_PASSWORD", "")
	viper.SetDefault("SMTP_FROM_EMAIL", "noreply@go-cms.com")
	viper.SetDefault("SMTP_FROM_NAME", "GO CMS")

	// CORS defaults
	viper.SetDefault("CORS_ALLOWED_ORIGINS", []string{"*"})
	viper.SetDefault("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"})
	viper.SetDefault("CORS_ALLOWED_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization"})
	viper.SetDefault("CORS_ALLOW_CREDENTIALS", true)

	// Rate limit defaults
	viper.SetDefault("RATE_LIMIT_REQUESTS", 100)
	viper.SetDefault("RATE_LIMIT_DURATION", "1m")

	// File upload defaults
	viper.SetDefault("MAX_UPLOAD_SIZE", 10485760) // 10MB
	viper.SetDefault("ALLOWED_FILE_TYPES", []string{"jpg", "jpeg", "png", "gif", "pdf", "docx", "xlsx", "mp4"})

	// Logging defaults
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_FORMAT", "json")
	viper.SetDefault("LOG_OUTPUT", "stdout")

	// Pagination defaults
	viper.SetDefault("DEFAULT_PAGE_SIZE", 20)
	viper.SetDefault("MAX_PAGE_SIZE", 100)
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetRedisAddr returns the Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetServerAddr returns the server address
func (c *ServerConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
