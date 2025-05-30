package dto

import (
	"encoding/json"
	"fmt"
)

type GameMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type CreateRoomPayload struct {
	Bet int `json:"bet"`
}
type JoinRoomPayload struct {
	RoomID string `json:"room_id"`
	Bet    int    `json:"bet"`
}

type LeaveRoomPayload struct{}

type ReadyPayload struct {
	IsReady bool `json:"is_ready"`
}

type HitPayload struct{}

type StandPayload struct{}

func MapToStruct(data interface{}, result interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data for struct conversion: %w", err)
	}
	if err := json.Unmarshal(jsonData, result); err != nil {
		return fmt.Errorf("failed to unmarshal data into struct: %w", err)
	}
	return nil
}
