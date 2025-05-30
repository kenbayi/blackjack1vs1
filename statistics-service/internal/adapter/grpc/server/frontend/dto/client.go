package dto

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	statisticsv1 "statistics/internal/adapter/grpc/server/frontend/proto"
	"statistics/internal/model"
)

// FromModelGeneralGameStatsToProto maps your domain model to the proto response.
func FromModelGeneralGameStatsToProto(stats model.GeneralGameStats) *statisticsv1.GetGeneralGameStatsResponse {
	return &statisticsv1.GetGeneralGameStatsResponse{
		Stats: &statisticsv1.GeneralGameStats{
			TotalUsers:       stats.TotalUsers,
			TotalGamesPlayed: stats.TotalGamesPlayed,
			TotalBetAmount:   stats.TotalBetAmount,
			LastUpdatedAt:    timestamppb.New(stats.LastUpdatedAt),
		},
	}
}

// FromModelUserGameStatsToProto maps your domain model to the proto response.
func FromModelUserGameStatsToProto(stats model.UserGameStats) *statisticsv1.GetUserGameStatsResponse {
	return &statisticsv1.GetUserGameStatsResponse{
		Stats: &statisticsv1.UserGameStats{
			UserId:           stats.UserID,
			GamesPlayed:      stats.GamesPlayed,
			GamesWon:         stats.GamesWon,
			GamesLost:        stats.GamesLost,
			GamesDrawn:       stats.GamesDrawn,
			TotalBet:         stats.TotalBet,
			TotalWinnings:    stats.TotalWinnings,
			TotalLosses:      stats.TotalLosses,
			WinRate:          stats.WinRate,  // float64 maps to double
			LossRate:         stats.LossRate, // float64 maps to double
			WinStreak:        stats.WinStreak,
			LossStreak:       stats.LossStreak,
			LastGamePlayedAt: timestamppb.New(stats.LastGamePlayedAt),
		},
	}
}

// FromModelLeaderboardToProto maps your domain model to the proto response.
func FromModelLeaderboardToProto(lb model.Leaderboard) *statisticsv1.GetLeaderboardResponse {
	protoEntries := make([]*statisticsv1.LeaderboardEntry, len(lb.Entries))
	for i, entry := range lb.Entries {
		protoEntries[i] = &statisticsv1.LeaderboardEntry{
			UserId:   entry.UserID,
			Username: entry.Username,
			Score:    entry.Score,
			Rank:     int32(entry.Rank),
		}
	}
	return &statisticsv1.GetLeaderboardResponse{
		Leaderboard: &statisticsv1.Leaderboard{
			Type:    lb.Type,
			Entries: protoEntries,
		},
	}
}
