package frontend

import (
	"context"
	"statistics/internal/model"
)

type StatisticsUseCase interface {
	GetGeneralGameStats(ctx context.Context) (model.GeneralGameStats, error)
	GetUserGameStats(ctx context.Context, userID int64) (model.UserGameStats, error)
	GetLeaderboard(ctx context.Context, leaderboardType string, limit int) (model.Leaderboard, error)
}
