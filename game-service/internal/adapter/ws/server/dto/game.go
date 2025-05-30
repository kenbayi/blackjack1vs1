package dto

import (
	"fmt"
	"game_svc/internal/model"
)

func FromCreateRequestToParams(payload CreateRoomPayload, userID string) *model.CreateRoomParams {
	return &model.CreateRoomParams{
		UserID: userID,
		Bet:    payload.Bet,
	}
}

func FromJoinRequestToParams(payload JoinRoomPayload, userID string) *model.JoinRoomParams {
	return &model.JoinRoomParams{
		UserID: userID,
		RoomID: payload.RoomID,
		Bet:    payload.Bet,
	}
}
func FromLeaveRequestToParams(roomID string, userID string) *model.LeaveRoomParams {
	return &model.LeaveRoomParams{
		UserID: userID,
		RoomID: roomID,
	}
}

func FromLeaveRequestToPlayerLeftNotification(room *model.Room, leftPlayerID string) *PlayerLeftNotificationDTO {
	playerIDs := make([]string, 0, len(room.Players))
	for _, p := range room.Players {
		if p.ID != leftPlayerID {
			playerIDs = append(playerIDs, p.ID)
		}
	}
	return &PlayerLeftNotificationDTO{
		Players: playerIDs,
		Message: fmt.Sprintf("Player %s has left the room.", leftPlayerID),
	}
}

func FromLeaveRequestToUpdateList(roomID string, room *model.Room, userID string, wasRoomDeleted bool) *RoomListUpdateDTO {
	if wasRoomDeleted {
		return &RoomListUpdateDTO{
			Action: "remove",
			RoomID: roomID,
		}
	}

	if room == nil {
		return &RoomListUpdateDTO{Action: "remove", RoomID: roomID}
	}

	playerIDs := make([]string, 0, len(room.Players))
	for _, p := range room.Players {
		if p.ID != userID {
			playerIDs = append(playerIDs, p.ID)
		}
	}
	return &RoomListUpdateDTO{
		Action:  "leave",
		RoomID:  room.ID,
		Status:  room.Status,
		Players: playerIDs,
		Bet:     room.Bet,
	}
}

func FromRoomModelToGameStateUpdate(room *model.Room, message string) *GameStateUpdate {
	if room == nil {
		return nil
	}

	playerHands := make(map[string][]string)
	playerScores := make(map[string]int)
	for _, p := range room.Players {
		handStr := make([]string, len(p.Hand))
		for i, card := range p.Hand {
			handStr[i] = card.Value + card.Suit
		}
		playerHands[p.ID] = handStr
		playerScores[p.ID] = p.Score
	}

	statePayload := map[string]interface{}{
		"hands":  playerHands,
		"scores": playerScores,
		"turn":   room.CurrentTurnPlayerID,
		"status": room.Status,
		"bet":    room.Bet,
	}

	return &GameStateUpdate{
		RoomID:  room.ID,
		State:   statePayload,
		Message: message,
	}
}

func MapModelHandsToStringHandsForAPI(modelHands map[string][]model.Card) map[string][]string {
	if modelHands == nil {
		return nil
	}

	stringHands := make(map[string][]string)
	for playerID, hand := range modelHands {
		sHand := make([]string, len(hand))
		for i, card := range hand {
			sHand[i] = cardModelToString(card)
		}
		stringHands[playerID] = sHand
	}
	return stringHands
}
func cardModelToString(card model.Card) string {
	return card.Value + card.Suit
}

// GetPlayerIDsFromModels извлекает срез ID игроков из среза []*model.Player.
func GetPlayerIDsFromModels(players []*model.Player) []string {
	if players == nil {
		return []string{} // Возвращаем пустой срез, если нет игроков
	}
	ids := make([]string, 0, len(players))
	for _, p := range players {
		if p != nil { // Дополнительная проверка на nil игрока в срезе
			ids = append(ids, p.ID)
		}
	}
	return ids
}

// RoomInfo для информации о комнате в списке
type RoomInfo struct {
	RoomID  string   `json:"roomID"`
	Status  string   `json:"status"`
	Players []string `json:"players"`
	Bet     int      `json:"bet"`
}

// CreateRoomResponse содержит данные для ответа создателю и для общего оповещения
type CreateRoomResponse struct {
	RoomID  string   `json:"roomID"`
	Action  string   `json:"action"`
	Status  string   `json:"status"`
	Players []string `json:"players"`
	Bet     int      `json:"bet"`
}

// JoinRoomResponse содержит данные для ответа присоединившемуся и для оповещения других
type JoinRoomResponse struct {
	Message string   `json:"msg"`
	Action  string   `json:"action"`
	RoomID  string   `json:"roomID"`
	Players []string `json:"players"`
}

// LeaveRoomResponse содержит данные для оповещения об уходе игрока
type LeaveRoomResponse struct {
	RoomID           string   `json:"roomID"`
	LeftPlayerID     string   `json:"leftPlayerID"`
	RemainingPlayers []string `json:"remainingPlayers"`
	IsRoomDeleted    bool     `json:"isRoomDeleted"` // Если комната удалена после ухода последнего игрока
}

// PlayerReadyResponse содержит данные для оповещения о готовности и возможном старте игры
type PlayerReadyResponse struct {
	RoomID                     string      `json:"roomID"`
	PlayerID                   string      `json:"playerID"` // Игрок, изменивший статус
	IsReady                    bool        `json:"isReady"`  // Его новый статус
	AllPlayersReady            bool        `json:"allPlayersReady"`
	GameStarted                bool        `json:"gameStarted"` // Флаг, что игра началась
	InitialGameStateIfStarted  interface{} `json:"initialGameStateIfStarted,omitempty"`
	RoomStatusIfGameStarted    string      `json:"roomStatusIfGameStarted,omitempty"`
	TurnIfGameStarted          string      `json:"turnIfGameStarted,omitempty"`
	RoomRemovedFromListOnStart bool        `json:"roomRemovedFromListOnStart"` // Для update_list action:remove
}

// HitResponse содержит данные об исходе взятия карты
type HitResponse struct {
	RoomID      string      `json:"roomID"`
	PlayerID    string      `json:"playerID"` // Игрок, который взял карту
	DealtCard   string      `json:"dealtCard"`
	NewScore    int         `json:"newScore"`
	IsBusted    bool        `json:"isBusted"`
	GameEnded   bool        `json:"gameEnded"`        // Если игра закончилась из-за bust
	Winner      string      `json:"winner,omitempty"` // Если игра закончилась
	FinalScores interface{} `json:"finalScores,omitempty"`
	FinalHands  interface{} `json:"finalHands,omitempty"`
	NextTurn    string      `json:"nextTurn,omitempty"` // Чей ход, если игра не закончилась
}

// StandResponse содержит данные об исходе решения "стоять"
type StandResponse struct {
	RoomID       string      `json:"roomID"`
	PlayerID     string      `json:"playerID"` // Игрок, который решил стоять
	CurrentScore int         `json:"currentScore"`
	GameEnded    bool        `json:"gameEnded"`        // Если игра закончилась (например, оба "стоят")
	Winner       string      `json:"winner,omitempty"` // Если игра закончилась
	FinalScores  interface{} `json:"finalScores,omitempty"`
	FinalHands   interface{} `json:"finalHands,omitempty"`
	NextTurn     string      `json:"nextTurn,omitempty"` // Чей ход, если игра не закончилась
}

// GameEndData содержит данные для сообщения game_end и game_waiting
type GameEndData struct {
	RoomID  string      `json:"roomID"`
	Winner  string      `json:"winner"`
	Loser   string      `json:"loser"`
	Scores  interface{} `json:"scores"`            // map[string]int
	Hands   interface{} `json:"hands"`             // map[string][]string
	Message string      `json:"message,omitempty"` // Для game_waiting
}

// DisconnectResponse содержит данные для оповещения об отключении игрока
type DisconnectResponse struct {
	LeftPlayerID           string          `json:"leftPlayerID"` // ID отключившегося игрока
	RoomID                 string          `json:"roomID"`
	RemainingPlayersInRoom []*model.Player `json:"remainingPlayersInRoom,omitempty"` // Модели оставшихся игроков
	UpdatedGameState       interface{}     `json:"updatedGameState,omitempty"`       // Обновленное состояние игры, если она продолжается или изменилась (структура зависит от фронтенда)
	GameEnded              bool            `json:"gameEnded"`                        // Завершилась ли игра из-за этого
	GameEndData            *GameEndData    `json:"gameEndData,omitempty"`            // Данные для game_end, если игра завершилась
	RoomRemovedFromList    bool            `json:"roomRemovedFromList"`              // Нужно ли обновить глобальный список комнат
	IsRoomDeleted          bool            `json:"isRoomDeleted"`                    // Была ли комната полностью удалена из хранилища
}
