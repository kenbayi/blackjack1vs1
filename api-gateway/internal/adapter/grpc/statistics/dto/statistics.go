package dto

import (
	svc "api-gateway/internal/adapter/frontend/proto/statistics"
	"api-gateway/internal/model"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func FromGRPCGeneralGameStatsResponse(resp *svc.GetGeneralGameStatsResponse) *model.GeneralGameStats {
	return &model.GeneralGameStats{
		TotalUsers:       resp.Stats.TotalUsers,
		TotalGamesPlayed: resp.Stats.TotalGamesPlayed,
		TotalBetAmount:   resp.Stats.TotalBetAmount,
		LastUpdatedAt:    *ProtoTimestampToTimePtr(resp.Stats.LastUpdatedAt),
	}
}

func FromGRPCUserGameStatsResponse(resp *svc.GetUserGameStatsResponse) *model.UserGameStats {
	return &model.UserGameStats{
		UserID:           resp.Stats.UserId,
		GamesPlayed:      resp.Stats.GamesPlayed,
		GamesWon:         resp.Stats.GamesWon,
		GamesLost:        resp.Stats.GamesLost,
		GamesDrawn:       resp.Stats.GamesDrawn,
		TotalBet:         resp.Stats.TotalBet,
		TotalWinnings:    resp.Stats.TotalWinnings,
		TotalLosses:      resp.Stats.TotalLosses,
		WinStreak:        resp.Stats.WinStreak,
		LossStreak:       resp.Stats.LossStreak,
		WinRate:          resp.Stats.WinRate,
		LossRate:         resp.Stats.LossRate,
		LastGamePlayedAt: *ProtoTimestampToTimePtr(resp.Stats.LastGamePlayedAt),
	}
}
func FromGRPCLeaderboardResponse(resp *svc.GetLeaderboardResponse) *model.Leaderboard {
	modelEntries := make([]model.LeaderboardEntry, 0, len(resp.Leaderboard.Entries))

	for _, grpcEntry := range resp.Leaderboard.Entries {
		if grpcEntry == nil {
			continue
		}

		modelEntries = append(modelEntries, model.LeaderboardEntry{
			UserID:   grpcEntry.UserId,
			Username: grpcEntry.Username,
			Score:    grpcEntry.Score,
			Rank:     int(grpcEntry.Rank),
		})
	}

	return &model.Leaderboard{
		Type:    resp.Leaderboard.Type,
		Entries: modelEntries,
	}
}

func ProtoTimestampToTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}
