package dao

import (
	"reflect"
	"statistics/internal/model"
	"strings"
	"time"
)

// GeneralStatsDAO represents the BSON structure for general statistics.
type GeneralStatsDAO struct {
	ID               string    `bson:"_id"` // Using string ID as in getGeneralStatsDocID()
	TotalUsers       int64     `bson:"total_users"`
	TotalGamesPlayed int64     `bson:"total_games_played"`
	TotalBetAmount   int64     `bson:"total_bet_amount"`
	LastUpdatedAt    time.Time `bson:"last_updated_at"`
}

// UserGameStatsDAO represents the BSON structure for user-specific game statistics.
type UserGameStatsDAO struct {
	// ID primitive.ObjectID `bson:"_id,omitempty"` // If using auto-generated MongoDB ObjectIDs
	UserID        int64 `bson:"user_id"` // This will be the query field, can also be _id
	GamesPlayed   int64 `bson:"games_played"`
	GamesWon      int64 `bson:"games_won"`
	GamesLost     int64 `bson:"games_lost"`
	GamesDrawn    int64 `bson:"games_drawn"`
	TotalBet      int64 `bson:"total_bet"`
	TotalWinnings int64 `bson:"total_winnings"`
	TotalLosses   int64 `bson:"total_losses"`
	// WinRate and LossRate are typically calculated, not stored, or updated transactionally.
	// If you want to store them, add them here with bson tags.
	WinStreak        int64     `bson:"win_streak"`
	LossStreak       int64     `bson:"loss_streak"`
	LastGamePlayedAt time.Time `bson:"last_game_played_at"`
}

// --- Mapping Functions ---

// ToGeneralStatsModel maps from GeneralStatsDAO to model.GeneralGameStats.
func ToGeneralStatsModel(dao GeneralStatsDAO) model.GeneralGameStats {
	return model.GeneralGameStats{
		TotalUsers:       dao.TotalUsers,
		TotalGamesPlayed: dao.TotalGamesPlayed,
		TotalBetAmount:   dao.TotalBetAmount,
		LastUpdatedAt:    dao.LastUpdatedAt,
	}
}

// ToUserGameStatsModel maps from UserGameStatsDAO to model.UserGameStats.
func ToUserGameStatsModel(dao UserGameStatsDAO) model.UserGameStats {
	stats := model.UserGameStats{
		UserID:           dao.UserID,
		GamesPlayed:      dao.GamesPlayed,
		GamesWon:         dao.GamesWon,
		GamesLost:        dao.GamesLost,
		GamesDrawn:       dao.GamesDrawn,
		TotalBet:         dao.TotalBet,
		TotalWinnings:    dao.TotalWinnings,
		TotalLosses:      dao.TotalLosses,
		WinStreak:        dao.WinStreak,
		LossStreak:       dao.LossStreak,
		LastGamePlayedAt: dao.LastGamePlayedAt,
	}
	// Calculate rates after mapping core fields
	if stats.GamesPlayed > 0 {
		stats.WinRate = float64(stats.GamesWon) / float64(stats.GamesPlayed)
		stats.LossRate = float64(stats.GamesLost) / float64(stats.GamesPlayed)
	} else {
		stats.WinRate = 0.0
		stats.LossRate = 0.0
	}
	return stats
}

func (g GeneralStatsDAO) GetBSONFieldName(goFieldName string) string {
	t := reflect.TypeOf(g)
	f, _ := t.FieldByName(goFieldName)
	return strings.Split(f.Tag.Get("bson"), ",")[0]
}
