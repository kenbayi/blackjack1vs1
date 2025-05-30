package model

import (
	"time"
)

// GeneralGameStats holds overall statistics for the Blackjack game.
type GeneralGameStats struct {
	TotalUsers       int64
	TotalGamesPlayed int64
	TotalBetAmount   int64 // Sum of all bets
	LastUpdatedAt    time.Time
}

// UserGameStats holds statistics for a specific user.
type UserGameStats struct {
	UserID           int64
	GamesPlayed      int64
	GamesWon         int64
	GamesLost        int64
	GamesDrawn       int64
	TotalBet         int64
	TotalWinnings    int64 // Sum of bets won
	TotalLosses      int64 // Sum of bets lost
	WinRate          float64
	LossRate         float64
	WinStreak        int64
	LossStreak       int64
	LastGamePlayedAt time.Time
}

// LeaderboardEntry represents an entry in a leaderboard.
type LeaderboardEntry struct {
	UserID   int64
	Username string // Optional
	Score    int64
	Rank     int
}

// Leaderboard represents a list of top players.
type Leaderboard struct {
	Type    string
	Entries []LeaderboardEntry
}

// GameHistory represents a single recorded game event.
type GameHistory struct {
	ID           string
	RoomID       string
	Player1ID    int64
	Player2ID    int64
	WinnerID     int64 // 0 for a draw
	LoserID      int64 // 0 for a draw (can be inferred if not a draw)
	BetAmount    int64
	Player1Hand  []string // String representation of cards, e.g., ["AH", "KD"]
	Player2Hand  []string
	Player1Score int32
	Player2Score int32
	GameEndedAt  time.Time
	// GameDuration time.Duration // Optional
}

// UserCreatedEventData holds the data for a user creation event.
type UserCreatedEventData struct {
	ID        int64
	Email     string
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
	IsDeleted bool
}

// UserDeletedEventData holds the data for a user deletion event.
type UserDeletedEventData struct {
	ID int64
}

// PlayerGameResultData holds data for a single player's game result.
type PlayerGameResultData struct {
	PlayerID   int64
	FinalScore int32
	FinalHand  []string
}

// GameResultEventData holds the data for a game result event.
type GameResultEventData struct {
	RoomID    string
	WinnerID  int64 // 0 for a draw
	LoserID   int64 // 0 for a draw
	Bet       int64
	CreatedAt time.Time // Timestamp of game end / event creation
	Player1   PlayerGameResultData
	Player2   PlayerGameResultData
}
