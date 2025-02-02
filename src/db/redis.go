package db

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
	Ctx         context.Context // Global context for Redis operations
)

func InitRedis(addr, password string) {
	// Initialize global context
	Ctx = context.Background()

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(Ctx, 5*time.Second) // Use a timeout for testing connection
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis!")
}
