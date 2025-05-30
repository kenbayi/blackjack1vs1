package frontend

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"statistics/internal/adapter/grpc/server/frontend/dto"
	statisticsv1 "statistics/internal/adapter/grpc/server/frontend/proto"
)

// StatisticsServer implements the gRPC StatisticsService.
type StatisticsServer struct {
	statisticsv1.UnimplementedStatisticsServiceServer
	uc StatisticsUseCase
}

// NewStatisticsServer creates a new StatisticsServer.
func NewStatisticsServer(uc StatisticsUseCase) *StatisticsServer {
	return &StatisticsServer{uc: uc}
}

func (s *StatisticsServer) GetGeneralGameStats(ctx context.Context, req *statisticsv1.GetGeneralGameStatsRequest) (*statisticsv1.GetGeneralGameStatsResponse, error) {
	domainStats, err := s.uc.GetGeneralGameStats(ctx)
	if err != nil {
		log.Printf("gRPC GetGeneralGameStats: Error from use case: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get general game stats: %v", err)
	}
	return dto.FromModelGeneralGameStatsToProto(domainStats), nil
}

func (s *StatisticsServer) GetUserGameStats(ctx context.Context, req *statisticsv1.GetUserGameStatsRequest) (*statisticsv1.GetUserGameStatsResponse, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required and cannot be zero")
	}

	domainStats, err := s.uc.GetUserGameStats(ctx, req.GetUserId())
	if err != nil {
		log.Printf("gRPC GetUserGameStats: Error from use case for UserID %d: %v", req.GetUserId(), err)
		return nil, status.Errorf(codes.Internal, "failed to get user game stats: %v", err)
	}
	return dto.FromModelUserGameStatsToProto(domainStats), nil
}

func (s *StatisticsServer) GetLeaderboard(ctx context.Context, req *statisticsv1.GetLeaderboardRequest) (*statisticsv1.GetLeaderboardResponse, error) {
	if req.GetLeaderboardType() == "" {
		return nil, status.Error(codes.InvalidArgument, "leaderboard_type is required")
	}
	if req.GetLimit() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "limit must be positive")
	}

	domainLeaderboard, err := s.uc.GetLeaderboard(ctx, req.GetLeaderboardType(), int(req.GetLimit())) // Convert limit to int
	if err != nil {
		log.Printf("gRPC GetLeaderboard: Error from use case for Type %s: %v", req.GetLeaderboardType(), err)
		return nil, status.Errorf(codes.Internal, "failed to get leaderboard: %v", err)
	}
	return dto.FromModelLeaderboardToProto(domainLeaderboard), nil
}
