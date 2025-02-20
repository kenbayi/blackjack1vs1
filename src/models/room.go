package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type GameRoom struct {
	ID        int       `json:"id"`         // Unique identifier
	RoomID    string    `json:"room_id"`    // UUID for the room
	Player1ID int       `json:"player1_id"` // Player 1's ID
	Player2ID int       `json:"player2_id"` // Player 2's ID
	Status    string    `json:"status"`     // Room status: "finished"
	Winner    string    `json:"winner"`
	CreatedAt time.Time `json:"created_at"` // Room creation timestamp
}

func InsertGameRoom(db *sql.DB, ctx context.Context, roomID string, p1, p2 string, winnerID string) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO game_rooms (room_id, player1_id, player2_id, status, winner, created_at)
		VALUES ($1, $2, $3, 'finished', $4, $5)`,
		roomID, p1, p2, winnerID, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("error inserting game record: %v", err)
	}
	return nil
}

func GetHistory(db *sql.DB, userID string) (*GameRoom, error) {
	id, err := strconv.Atoi(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid playerID: %v", err)
	}
	query := `SELECT room_id, player1_id, player2_id, status, winner, created_at FROM game_rooms WHERE player1_id = $1 OR player2_id = $1`

	var stats GameRoom
	err = db.QueryRow(query, id).Scan(&stats.RoomID, &stats.Player1ID, &stats.Player2ID, &stats.Status, &stats.Winner, &stats.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // History not found
	} else if err != nil {
		return nil, err
	}
	return &stats, nil
}
