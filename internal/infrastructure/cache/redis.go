package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/config"
	"github.com/owner/go-cms/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Client is the global Redis client
var Client *redis.Client

// Init initializes the Redis connection
func Init(cfg *config.RedisConfig) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connected successfully",
		zap.String("addr", cfg.GetRedisAddr()),
		zap.Int("db", cfg.DB),
	)

	return nil
}

// Close closes the Redis connection
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// GetClient returns the Redis client
func GetClient() *redis.Client {
	return Client
}

// Set sets a key-value pair with expiration
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return Client.Set(ctx, key, value, expiration).Err()
}

// Get gets a value by key
func Get(ctx context.Context, key string) (string, error) {
	return Client.Get(ctx, key).Result()
}

// Delete deletes a key
func Delete(ctx context.Context, keys ...string) error {
	return Client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func Exists(ctx context.Context, keys ...string) (int64, error) {
	return Client.Exists(ctx, keys...).Result()
}

// Expire sets an expiration on a key
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return Client.Expire(ctx, key, expiration).Err()
}

// HSet sets a hash field
func HSet(ctx context.Context, key string, values ...interface{}) error {
	return Client.HSet(ctx, key, values...).Err()
}

// HGet gets a hash field
func HGet(ctx context.Context, key, field string) (string, error) {
	return Client.HGet(ctx, key, field).Result()
}

// HGetAll gets all hash fields
func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return Client.HGetAll(ctx, key).Result()
}

// HDel deletes hash fields
func HDel(ctx context.Context, key string, fields ...string) error {
	return Client.HDel(ctx, key, fields...).Err()
}

// Incr increments a key
func Incr(ctx context.Context, key string) (int64, error) {
	return Client.Incr(ctx, key).Result()
}

// Decr decrements a key
func Decr(ctx context.Context, key string) (int64, error) {
	return Client.Decr(ctx, key).Result()
}

// SetNX sets a key only if it doesn't exist
func SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return Client.SetNX(ctx, key, value, expiration).Result()
}

// GetDel gets and deletes a key
func GetDel(ctx context.Context, key string) (string, error) {
	return Client.GetDel(ctx, key).Result()
}

// Keys returns all keys matching pattern
func Keys(ctx context.Context, pattern string) ([]string, error) {
	return Client.Keys(ctx, pattern).Result()
}

// FlushDB flushes the current database
func FlushDB(ctx context.Context) error {
	return Client.FlushDB(ctx).Err()
}

// Cache key prefixes
const (
	PrefixUser       = "user:"
	PrefixSession    = "session:"
	PrefixOTP        = "otp:"
	PrefixPermission = "permission:"
	PrefixRateLimit  = "ratelimit:"
	PrefixCache      = "cache:"
)

// BuildKey builds a cache key with prefix
func BuildKey(prefix, id string) string {
	return prefix + id
}

// CacheUserSession caches a user session
func CacheUserSession(ctx context.Context, sessionID string, userID uuid.UUID, expiration time.Duration) error {
	key := BuildKey(PrefixSession, sessionID)
	return Set(ctx, key, userID.String(), expiration)
}

// GetUserSession gets a user session
func GetUserSession(ctx context.Context, sessionID string) (string, error) {
	key := BuildKey(PrefixSession, sessionID)
	return Get(ctx, key)
}

// DeleteUserSession deletes a user session
func DeleteUserSession(ctx context.Context, sessionID string) error {
	key := BuildKey(PrefixSession, sessionID)
	return Delete(ctx, key)
}

// CacheOTP caches an OTP
func CacheOTP(ctx context.Context, email, otp string, expiration time.Duration) error {
	key := BuildKey(PrefixOTP, email)
	return Set(ctx, key, otp, expiration)
}

// GetOTP gets an OTP
func GetOTP(ctx context.Context, email string) (string, error) {
	key := BuildKey(PrefixOTP, email)
	return Get(ctx, key)
}

// DeleteOTP deletes an OTP
func DeleteOTP(ctx context.Context, email string) error {
	key := BuildKey(PrefixOTP, email)
	return Delete(ctx, key)
}

// CachePermissions caches user permissions
func CachePermissions(ctx context.Context, userID uuid.UUID, permissions []string, expiration time.Duration) error {
	key := BuildKey(PrefixPermission, userID.String())
	// Store as hash for efficient access
	values := make([]interface{}, 0, len(permissions)*2)
	for _, perm := range permissions {
		values = append(values, perm, "1")
	}
	if err := HSet(ctx, key, values...); err != nil {
		return err
	}
	return Expire(ctx, key, expiration)
}

// GetPermissions gets cached user permissions
func GetPermissions(ctx context.Context, userID uuid.UUID) (map[string]string, error) {
	key := BuildKey(PrefixPermission, userID.String())
	return HGetAll(ctx, key)
}

// InvalidatePermissions invalidates user permissions cache
func InvalidatePermissions(ctx context.Context, userID uuid.UUID) error {
	key := BuildKey(PrefixPermission, userID.String())
	return Delete(ctx, key)
}

// CheckRateLimit checks rate limit for a key
func CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	rateLimitKey := BuildKey(PrefixRateLimit, key)

	count, err := Incr(ctx, rateLimitKey)
	if err != nil {
		return false, err
	}

	if count == 1 {
		// First request, set expiration
		if err := Expire(ctx, rateLimitKey, window); err != nil {
			return false, err
		}
	}

	return count <= limit, nil
}
