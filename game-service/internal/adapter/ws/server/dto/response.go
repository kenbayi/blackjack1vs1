package dto

import "game_svc/internal/model"

type ErrorResponse struct {
	ErrorType string `json:"error_type"`
	Message   string `json:"message"`
}

type RoomCreatedResponse struct {
	RoomID  string `json:"room_id"`
	UserID  string `json:"user_id"`
	Message string `json:"message,omitempty"`
}

type PlayerLeftNotificationDTO struct {
	RoomID  string   `json:"roomID"`
	Players []string `json:"players"`
	Message string   `json:"message,omitempty"`
}

type RoomListUpdateDTO struct {
	Action  string   `json:"action"`
	RoomID  string   `json:"roomID"`
	Status  string   `json:"status,omitempty"`
	Players []string `json:"players,omitempty"`
	Bet     int      `json:"bet,omitempty"`
}

type PlayerLeftNotification struct {
	Action  string   `json:"action"`
	RoomID  string   `json:"room_id"`
	UserID  string   `json:"user_id"`
	Message string   `json:"message,omitempty"`
	Players []string `json:"players"`
	Status  string   `json:"status"`
	Bet     int      `json:"bet"`
}

type GameStateUpdate struct {
	RoomID  string      `json:"room_id"`
	State   interface{} `json:"state"`
	Message string      `json:"message,omitempty"`
}

// HitBroadcastPayloadDTO - для сообщения "hit"
type HitBroadcastPayloadDTO struct {
	ForPlayer string `json:"forPlayer"`
	Card      string `json:"card"`
	Score     int    `json:"score"`
}

// BustedBroadcastPayloadDTO - для сообщения "busted"
type BustedBroadcastPayloadDTO struct {
	ForPlayer string `json:"forPlayer"`
	Message   string `json:"msg"`
}

// GameEndBroadcastPayloadDTO - для сообщения "game_end"
type GameEndBroadcastPayloadDTO struct {
	RoomID string              `json:"roomID"`
	Winner string              `json:"winner"`
	Scores map[string]int      `json:"scores"`
	Hands  map[string][]string `json:"hands"` // Руки как строки карт
}

// TurnBroadcastPayloadDTO - для сообщения "turn"
type TurnBroadcastPayloadDTO struct {
	Turn string `json:"turn"`
}

func FromModelToListResponse(ucResponse *model.Room) *CreateRoomResponse {
	playerIDs := make([]string, 0, len(ucResponse.Players))
	for _, p := range ucResponse.Players {
		playerIDs = append(playerIDs, p.ID)
	}
	return &CreateRoomResponse{
		RoomID:  ucResponse.ID,
		Action:  "create",
		Status:  ucResponse.Status,
		Players: playerIDs,
		Bet:     ucResponse.Bet,
	}
}
