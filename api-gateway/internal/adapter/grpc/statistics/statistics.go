package statistics

import (
	svc "api-gateway/internal/adapter/frontend/proto/statistics"
	"api-gateway/internal/adapter/grpc/statistics/dto"
	"api-gateway/internal/model"
	"context"
)

type Statistics struct {
	statistics svc.StatisticsServiceClient
}

func NewStatistics(statistics svc.StatisticsServiceClient) *Statistics {
	return &Statistics{
		statistics: statistics,
	}
}

func (c *Statistics) GetGeneralGameStats(ctx context.Context) (*model.GeneralGameStats, error) {
	resp, err := c.statistics.GetGeneralGameStats(ctx, &svc.GetGeneralGameStatsRequest{})
	if err != nil {
		return nil, err
	}
	return dto.FromGRPCGeneralGameStatsResponse(resp), nil
}

func (c *Statistics) GetUserGameStats(ctx context.Context, userID int64) (*model.UserGameStats, error) {
	resp, err := c.statistics.GetUserGameStats(ctx, &svc.GetUserGameStatsRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}
	return dto.FromGRPCUserGameStatsResponse(resp), nil
}

func (c *Statistics) GetLeaderboard(ctx context.Context, request model.Leaderboard) (*model.Leaderboard, error) {
	resp, err := c.statistics.GetLeaderboard(ctx, &svc.GetLeaderboardRequest{
		LeaderboardType: request.Type,
		Limit:           *request.Limit,
	})
	if err != nil {
		return nil, err
	}
	return dto.FromGRPCLeaderboardResponse(resp), nil

}
