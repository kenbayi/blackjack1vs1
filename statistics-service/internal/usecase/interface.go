package usecase

import (
	"context"
	"statistics/internal/model"
	"time"
)

// StatisticsRepository defines methods for persisting and retrieving statistics.
type StatisticsRepository interface {
	IncrementTotalUsers(ctx context.Context) error
	DecrementTotalUsers(ctx context.Context) error
	UpdateStatsForGameResult(ctx context.Context, gameResult model.GameResultEventData) error
	RepoGetGeneralGameStats(ctx context.Context) (model.GeneralGameStats, error)
	RepoGetUserGameStats(ctx context.Context, userID int64) (model.UserGameStats, error)
	RepoGetLeaderboard(ctx context.Context, leaderboardType string, limit int) (model.Leaderboard, error)
}

// StatisticsRedisCache defines methods for caching statistics in Redis.
type StatisticsRedisCache interface {
	RepoGetGeneralGameStats(ctx context.Context) (model.GeneralGameStats, error)
	SetGeneralGameStats(ctx context.Context, stats model.GeneralGameStats, ttl time.Duration) error
	DeleteGeneralGameStats(ctx context.Context) error
	RepoGetUserGameStats(ctx context.Context, userID int64) (model.UserGameStats, error)
	SetUserGameStats(ctx context.Context, userID int64, stats model.UserGameStats, ttl time.Duration) error
	DeleteUserGameStats(ctx context.Context, userID int64) error
	RepoGetLeaderboard(ctx context.Context, leaderboardType string, limit int) (model.Leaderboard, error)
	SetLeaderboard(ctx context.Context, leaderboardType string, leaderboard model.Leaderboard, ttl time.Duration) error
	DeleteLeaderboard(ctx context.Context, leaderboardType string) error
}

// GameHistoryRepository defines methods for persisting game history.
type GameHistoryRepository interface {
	InsertGame(ctx context.Context, gameHistoryEntry model.GameHistory) error
}
