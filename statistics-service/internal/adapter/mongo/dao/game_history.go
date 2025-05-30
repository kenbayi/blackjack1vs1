package dao

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"statistics/internal/model"
	"time"
)

// GameHistoryDAO represents the BSON structure for a game history entry.
type GameHistoryDAO struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"` // Auto-generated MongoDB ID
	RoomID       string             `bson:"room_id"`
	Player1ID    int64              `bson:"player1_id"`
	Player2ID    int64              `bson:"player2_id"`
	WinnerID     int64              `bson:"winner_id"` // 0 for a draw
	LoserID      int64              `bson:"loser_id"`  // 0 for a draw
	BetAmount    int64              `bson:"bet_amount"`
	Player1Hand  []string           `bson:"player1_hand"`
	Player2Hand  []string           `bson:"player2_hand"`
	Player1Score int32              `bson:"player1_score"`
	Player2Score int32              `bson:"player2_score"`
	GameEndedAt  time.Time          `bson:"game_ended_at"`
}

// FromGameHistoryModel maps model.GameHistory to GameHistoryDAO for storage.
func FromGameHistoryModel(m model.GameHistory) GameHistoryDAO {
	return GameHistoryDAO{
		RoomID:       m.RoomID,
		Player1ID:    m.Player1ID,
		Player2ID:    m.Player2ID,
		WinnerID:     m.WinnerID,
		LoserID:      m.LoserID,
		BetAmount:    m.BetAmount,
		Player1Hand:  m.Player1Hand,
		Player2Hand:  m.Player2Hand,
		Player1Score: m.Player1Score,
		Player2Score: m.Player2Score,
		GameEndedAt:  m.GameEndedAt,
	}
}
