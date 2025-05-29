package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"statistics/internal/model"
	"statistics/pkg/redis"
)

const (
	generalGameStatsKeyPrefix = "stats:blackjack:general"
	userGameStatsKeyPrefix    = "stats:blackjack:user:%d"        // %d for int64 userID
	leaderboardKeyPrefix      = "stats:blackjack:leaderboard:%s" // %s for leaderboardType
)

type StatisticsRedisCacheImpl struct {
	client *redis.Client
}

func NewStatisticsRedisCache(client *redis.Client) *StatisticsRedisCacheImpl {
	return &StatisticsRedisCacheImpl{
		client: client,
	}
}

// --- GeneralGameStats Cache Methods ---

func (r *StatisticsRedisCacheImpl) RepoGetGeneralGameStats(ctx context.Context) (model.GeneralGameStats, error) {
	key := generalGameStatsKeyPrefix
	data, err := r.client.Unwrap().Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return model.GeneralGameStats{}, fmt.Errorf("general game stats not found in cache: %w", err) // Or return a specific "cache miss" error
		}
		return model.GeneralGameStats{}, fmt.Errorf("redis Get for general game stats failed: %w", err)
	}

	var stats model.GeneralGameStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return model.GeneralGameStats{}, fmt.Errorf("failed to unmarshal general game stats from cache: %w", err)
	}
	return stats, nil
}

func (r *StatisticsRedisCacheImpl) SetGeneralGameStats(ctx context.Context, stats model.GeneralGameStats, ttl time.Duration) error {
	key := generalGameStatsKeyPrefix
	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal general game stats for cache: %w", err)
	}

	if err := r.client.Unwrap().Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis Set for general game stats failed: %w", err)
	}
	return nil
}

func (r *StatisticsRedisCacheImpl) DeleteGeneralGameStats(ctx context.Context) error {
	key := generalGameStatsKeyPrefix
	if err := r.client.Unwrap().Del(ctx, key).Err(); err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil
		}
		return fmt.Errorf("redis Del for general game stats failed: %w", err)
	}
	return nil
}

// --- UserGameStats Cache Methods ---

func userStatsKey(userID int64) string {
	return fmt.Sprintf(userGameStatsKeyPrefix, userID)
}

func (r *StatisticsRedisCacheImpl) RepoGetUserGameStats(ctx context.Context, userID int64) (model.UserGameStats, error) {
	key := userStatsKey(userID)
	data, err := r.client.Unwrap().Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return model.UserGameStats{}, fmt.Errorf("user game stats for userID %d not found in cache: %w", userID, err) // Cache miss
		}
		return model.UserGameStats{}, fmt.Errorf("redis Get for user game stats (userID %d) failed: %w", userID, err)
	}

	var stats model.UserGameStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return model.UserGameStats{}, fmt.Errorf("failed to unmarshal user game stats from cache (userID %d): %w", userID, err)
	}
	return stats, nil
}

func (r *StatisticsRedisCacheImpl) SetUserGameStats(ctx context.Context, userID int64, stats model.UserGameStats, ttl time.Duration) error {
	key := userStatsKey(userID)
	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal user game stats for cache (userID %d): %w", userID, err)
	}

	if err := r.client.Unwrap().Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis Set for user game stats (userID %d) failed: %w", userID, err)
	}
	return nil
}

func (r *StatisticsRedisCacheImpl) DeleteUserGameStats(ctx context.Context, userID int64) error {
	key := userStatsKey(userID)
	if err := r.client.Unwrap().Del(ctx, key).Err(); err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil // Key didn't exist, which is fine for a delete
		}
		return fmt.Errorf("redis Del for user game stats (userID %d) failed: %w", userID, err)
	}
	return nil
}

// --- Leaderboard Cache Methods ---

func leaderboardCacheKey(leaderboardType string) string {
	return fmt.Sprintf(leaderboardKeyPrefix, leaderboardType)
}

func (r *StatisticsRedisCacheImpl) RepoGetLeaderboard(ctx context.Context, leaderboardType string, limit int) (model.Leaderboard, error) {
	// For leaderboards, you might store the whole leaderboard object, or just the sorted set data.
	// If storing the whole object:
	key := leaderboardCacheKey(leaderboardType)
	data, err := r.client.Unwrap().Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return model.Leaderboard{}, fmt.Errorf("leaderboard %s not found in cache: %w", leaderboardType, err) // Cache miss
		}
		return model.Leaderboard{}, fmt.Errorf("redis Get for leaderboard %s failed: %w", leaderboardType, err)
	}

	var lb model.Leaderboard
	if err := json.Unmarshal(data, &lb); err != nil {
		return model.Leaderboard{}, fmt.Errorf("failed to unmarshal leaderboard %s from cache: %w", leaderboardType, err)
	}
	// Optionally, you could trim the lb.Entries to the requested limit here if the cached version is longer.
	// However, it's often better to cache for a specific limit or let the use case handle trimming.
	return lb, nil
}

func (r *StatisticsRedisCacheImpl) SetLeaderboard(ctx context.Context, leaderboardType string, leaderboard model.Leaderboard, ttl time.Duration) error {
	key := leaderboardCacheKey(leaderboardType)
	data, err := json.Marshal(leaderboard)
	if err != nil {
		return fmt.Errorf("failed to marshal leaderboard %s for cache: %w", leaderboardType, err)
	}

	if err := r.client.Unwrap().Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis Set for leaderboard %s failed: %w", leaderboardType, err)
	}
	return nil
}

func (r *StatisticsRedisCacheImpl) DeleteLeaderboard(ctx context.Context, leaderboardType string) error {
	key := leaderboardCacheKey(leaderboardType)
	if err := r.client.Unwrap().Del(ctx, key).Err(); err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil
		}
		return fmt.Errorf("redis Del for leaderboard %s failed: %w", leaderboardType, err)
	}
	return nil
}
