package models

import (
	"time"
)

type GameRoom struct {
	ID        int       `json:"id"`         // Unique identifier
	RoomID    string    `json:"room_id"`    // UUID for the room
	Player1ID int       `json:"player1_id"` // Player 1's ID
	Player2ID int       `json:"player2_id"` // Player 2's ID
	Status    string    `json:"status"`     // Room status: "waiting", "in-progress", "finished"
	CreatedAt time.Time `json:"created_at"` // Room creation timestamp
}
