package dto

import (
	"api-gateway/internal/model"
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type GeneralGameStatsResponse struct {
	TotalUsers       int64     `json:"total_users"`
	TotalGamesPlayed int64     `json:"total_games_played"`
	TotalBetAmount   int64     `json:"total_bet_amount"`
	LastUpdatedAt    time.Time `json:"last_updated_at"`
}

type UserGameStatsResponse struct {
	UserID           int64     `json:"user_id"`
	GamesPlayed      int64     `json:"games_played"`
	GamesWon         int64     `json:"games_won"`
	GamesLost        int64     `json:"games_lost"`
	GamesDrawn       int64     `json:"games_drawn"`
	TotalBet         int64     `json:"total_bet"`
	TotalWinnings    int64     `json:"total_winnings"`
	TotalLosses      int64     `json:"total_losses"`
	WinRate          float64   `json:"win_rate"`
	LossRate         float64   `json:"loss_rate"`
	WinStreak        int64     `json:"win_streak"`
	LossStreak       int64     `json:"loss_streak"`
	LastGamePlayedAt time.Time `json:"last_game_played_at"`
}

type LeaderboardEntryResponse struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username,omitempty"`
	Score    int64  `json:"score"`
	Rank     int    `json:"rank"`
}

type LeaderboardResponse struct {
	Type    string                     `json:"type"`
	Entries []LeaderboardEntryResponse `json:"entries"`
}

func FromModelToGeneralGameStatsResponse(stats model.GeneralGameStats) GeneralGameStatsResponse {
	return GeneralGameStatsResponse{
		TotalUsers:       stats.TotalUsers,
		TotalGamesPlayed: stats.TotalGamesPlayed,
		TotalBetAmount:   stats.TotalBetAmount,
		LastUpdatedAt:    stats.LastUpdatedAt,
	}
}

func ToUserGameStatsRequest(ctx *gin.Context) (int64, error) {
	userIDStr := ctx.Param("userID")
	if userIDStr == "" {
		return 0, errors.New("user id is required")
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, errors.New("invalid user id format")
	}
	return userID, nil
}

func FromModelToUserGameStatsResponse(stats model.UserGameStats) UserGameStatsResponse {
	return UserGameStatsResponse{
		UserID:           stats.UserID,
		GamesPlayed:      stats.GamesPlayed,
		GamesWon:         stats.GamesWon,
		GamesLost:        stats.GamesLost,
		GamesDrawn:       stats.GamesDrawn,
		TotalBet:         stats.TotalBet,
		TotalWinnings:    stats.TotalWinnings,
		TotalLosses:      stats.TotalLosses,
		WinRate:          stats.WinRate,
		LossRate:         stats.LossRate,
		WinStreak:        stats.WinStreak,
		LossStreak:       stats.LossStreak,
		LastGamePlayedAt: stats.LastGamePlayedAt,
	}
}

func ToLeaderboardRequest(ctx *gin.Context) (model.Leaderboard, error) {
	leaderboardType := ctx.Query("type")
	if leaderboardType == "" {
		return model.Leaderboard{}, errors.New("leaderboard type is required")
	}

	req := model.Leaderboard{
		Type: leaderboardType,
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		limit, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			return model.Leaderboard{}, errors.New("invalid limit format")
		}
		limit32 := int32(limit)
		req.Limit = &limit32
	}

	return req, nil
}

func FromModelToLeaderboardResponse(leaderboard model.Leaderboard) LeaderboardResponse {
	entries := make([]LeaderboardEntryResponse, 0, len(leaderboard.Entries))
	for _, entry := range leaderboard.Entries {
		entries = append(entries, LeaderboardEntryResponse{
			UserID:   entry.UserID,
			Username: entry.Username,
			Score:    entry.Score,
			Rank:     entry.Rank,
		})
	}
	return LeaderboardResponse{
		Type:    leaderboard.Type,
		Entries: entries,
	}
}
