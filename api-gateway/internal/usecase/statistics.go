package usecase

import (
	"context"

	"api-gateway/internal/model"
)

type Statistics struct {
	presenter StatisticsPresenter
}

func NewStatistics(p StatisticsPresenter) *Statistics {
	return &Statistics{presenter: p}
}

func (s *Statistics) GetGeneralGameStats(ctx context.Context) (*model.GeneralGameStats, error) {
	return s.presenter.GetGeneralGameStats(ctx)
}

func (s *Statistics) GetUserGameStats(ctx context.Context, userID int64) (*model.UserGameStats, error) {
	return s.presenter.GetUserGameStats(ctx, userID)
}

func (s *Statistics) GetLeaderboard(ctx context.Context, req model.Leaderboard) (*model.Leaderboard, error) {
	return s.presenter.GetLeaderboard(ctx, req)
}
