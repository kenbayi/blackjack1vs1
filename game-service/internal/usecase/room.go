package usecase

import (
	"context"
	"errors"
	"fmt"
	"game_svc/internal/model"
	"log"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// RoomServiceImpl реализует интерфейс RoomUseCase.
type RoomServiceImpl struct {
	roomStateRepo   RoomStateRepository
	clientPresenter ClientPresenter
}

// NewRoomService создает новый экземпляр RoomServiceImpl.
func NewRoomService(
	rsr RoomStateRepository,
	presenter ClientPresenter,
) *RoomServiceImpl {
	return &RoomServiceImpl{
		roomStateRepo:   rsr,
		clientPresenter: presenter,
	}
}

func generateRoomID() string {
	return uuid.New().String()
}

var playerSpecificBaseFields = []string{"readyStatus", "scores", "hands", "lastAction", "stood"}

// CreateRoom реализует логику создания комнаты.
func (s *RoomServiceImpl) CreateRoom(params model.CreateRoomParams) (*model.Room, error) {
	ctx := context.Background()
	userID := params.UserID
	bet := params.Bet
	userIDint, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, err
	}
	//1. Проверка баланса игрока
	playerBalance, err := s.clientPresenter.Get(ctx, userIDint)
	if err != nil {
		log.Printf("Error getting player balance for %s: %v", userID, err)
		return nil, fmt.Errorf("failed to get player balance: %w", err)
	}
	if int(*playerBalance.Balance) < bet {
		return nil, errors.New("insufficient funds to create a room")
	}

	// 2. Генерация ID комнаты
	roomID := generateRoomID()

	// 3. Создание доменной модели игрока-создателя
	creatorPlayer := &model.Player{
		ID:         userID,
		IsReady:    false,
		Score:      0,
		LastAction: "",
		Hand:       []model.Card{},
	}

	// 4. Создание доменной модели комнаты
	newRoom := &model.Room{
		ID:                  roomID,
		Status:              "waiting",
		Bet:                 bet,
		Players:             []*model.Player{creatorPlayer},
		Deck:                []model.Card{},
		CurrentTurnPlayerID: "",
	}
	err = s.roomStateRepo.SaveRoom(ctx, newRoom)

	if err != nil {
		log.Printf("Error saving room state to Redis for room %s: %v", newRoom.ID, err)
		return nil, fmt.Errorf("failed to save room state: %w", err)
	}

	log.Printf("Use Case: Room %s created successfully by user %s with bet %d", newRoom.ID, userID, bet)
	return newRoom, nil
}

// JoinRoom implements the logic for a player to join a room.
func (s *RoomServiceImpl) JoinRoom(params model.JoinRoomParams) (*model.Room, error) { // Assuming model.JoinRoomParams
	ctx := context.Background()
	joiningUserID := params.UserID
	roomID := params.RoomID
	clientBet := params.Bet

	// 1. Get current room state from Redis for validation
	roomStateMap, err := s.roomStateRepo.GetAllRoomFields(ctx, roomID)
	if err != nil {
		log.Printf("Use Case JoinRoom: Error retrieving room state for room %s: %v", roomID, err)
		return nil, fmt.Errorf("error retrieving room state: %w", err)
	}
	if len(roomStateMap) == 0 {
		return nil, errors.New("room not found")
	}

	roomStatus := roomStateMap["status"]
	roomBetStoredStr, okBet := roomStateMap["bet"]
	if !okBet {
		log.Printf("Use Case JoinRoom: Bet not found in Redis for room %s", roomID)
		return nil, errors.New("room configuration error: bet not set")
	}
	currentPlayersStr, okPlayers := roomStateMap["players"]
	if !okPlayers {
		log.Printf("Use Case JoinRoom: Players field not found in Redis for room %s", roomID)
		return nil, errors.New("room configuration error: players not set")
	}

	roomBetStored, err := strconv.Atoi(roomBetStoredStr)
	if err != nil {
		log.Printf("Use Case JoinRoom: Invalid bet format ('%s') in Redis for room %s: %v", roomBetStoredStr, roomID, err)
		return nil, errors.New("room has invalid bet configuration")
	}

	// 2. Validations
	existingPlayerIDs := splitPlayers(currentPlayersStr)

	if len(existingPlayerIDs) >= 2 {
		return nil, errors.New("room is full")
	}

	for _, pID := range existingPlayerIDs {
		if pID == joiningUserID {
			return nil, errors.New("player already in this room")
		}
	}

	if clientBet != roomBetStored {
		return nil, fmt.Errorf("your bet (%d) does not match the room bet (%d)", clientBet, roomBetStored)
	}
	joiningUserIDint, ok := strconv.ParseInt(joiningUserID, 10, 64)
	if ok != nil {
		return nil, fmt.Errorf("could not parse joining user id %s", joiningUserIDint)
	}
	playerBalance, err := s.clientPresenter.Get(ctx, joiningUserIDint)
	if err != nil {
		log.Printf("Use Case JoinRoom: Error getting player balance for %s: %v", joiningUserID, err)
		return nil, fmt.Errorf("failed to get player balance: %w", err)
	}
	if int(*playerBalance.Balance) < roomBetStored {
		return nil, errors.New("insufficient funds to join the room")
	}

	// 3. Update Redis using the new repository method
	newPlayerIDsForRedisList := append(existingPlayerIDs, joiningUserID)
	updatedPlayersStrRedis := strings.Join(newPlayerIDsForRedisList, ",")

	// Call the dedicated repository method
	if err := s.roomStateRepo.AddJoiningPlayer(ctx, roomID, joiningUserID, updatedPlayersStrRedis); err != nil {
		log.Printf("Use Case JoinRoom: Failed to add player %s to room %s via repository: %v", joiningUserID, roomID, err)
		return nil, fmt.Errorf("failed to update room state for joining player: %w", err)
	}

	// 4. Construct and return the updated domain model *model.Room
	finalPlayersInModel := make([]*model.Player, 0, len(newPlayerIDsForRedisList))
	for _, pID := range newPlayerIDsForRedisList {
		player := &model.Player{ID: pID}
		if pID == joiningUserID {
			player.IsReady = false
			player.Score = 0
			player.LastAction = ""
			player.Hand = []model.Card{}
		} else {
			readyStr := roomStateMap["readyStatus."+pID]
			player.IsReady = readyStr == "1"

			scoreStr := roomStateMap["scores."+pID]
			score, _ := strconv.Atoi(scoreStr)
			player.Score = score
			player.LastAction = roomStateMap["lastAction."+pID]
			handStr := roomStateMap["hand."+pID]
			player.Hand = parseHandString(handStr)
			// Potentially load Hand and LastAction from roomStateMap too if they exist
		}
		finalPlayersInModel = append(finalPlayersInModel, player)
	}

	updatedRoomModel := &model.Room{
		ID:                  roomID,
		Status:              roomStatus,
		Bet:                 roomBetStored,
		Players:             finalPlayersInModel,
		CurrentTurnPlayerID: roomStateMap["turn"],
		Deck:                []model.Card{},
	}

	log.Printf("Use Case: User %s joined room %s. Total players now: %d", joiningUserID, roomID, len(updatedRoomModel.Players))
	return updatedRoomModel, nil
}

// Вспомогательная функция splitPlayers остается той же
func splitPlayers(playersStr string) []string {
	if playersStr == "" {
		return []string{}
	}
	var result []string
	parts := strings.Split(playersStr, ",")
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func (s *RoomServiceImpl) LeaveRoom(params model.LeaveRoomParams) (updatedRoom *model.Room, wasRoomDeleted bool, err error) {
	ctx := context.Background()
	leavingUserID := params.UserID
	roomID := params.RoomID

	log.Printf("Use Case LeaveRoom: User %s attempting to leave room %s", leavingUserID, roomID)

	// 1. Получаем текущее состояние комнаты из Redis
	roomStateMap, err := s.roomStateRepo.GetAllRoomFields(ctx, roomID)
	if err != nil {
		log.Printf("Use Case LeaveRoom: Error retrieving room state for room %s: %v", roomID, err)
		return nil, false, fmt.Errorf("error retrieving room state: %w", err)
	}
	if len(roomStateMap) == 0 {
		log.Printf("Use Case LeaveRoom: Room %s not found in Redis for leave.", roomID)
		return nil, true, errors.New("room not found")
	}

	currentPlayersStr := roomStateMap["players"]
	initialPlayerIDs := splitPlayers(currentPlayersStr)

	// 2. Проверяем, есть ли игрок в комнате
	isPresent := false
	var remainingPlayerIDsAfterLeave []string
	for _, pID := range initialPlayerIDs {
		if pID == leavingUserID {
			isPresent = true
		} else {
			remainingPlayerIDsAfterLeave = append(remainingPlayerIDsAfterLeave, pID)
		}
	}

	if !isPresent {
		log.Printf("Use Case LeaveRoom: Player %s not found in room %s players list (%s)", leavingUserID, roomID, currentPlayersStr)
		return nil, false, errors.New("player not in this room")
	}

	// 3. Обновляем список игроков в Redis
	updatedPlayersStrRedis := strings.Join(remainingPlayerIDsAfterLeave, ",")
	if err := s.roomStateRepo.UpdatePlayerList(ctx, roomID, updatedPlayersStrRedis); err != nil {
		log.Printf("Use Case LeaveRoom: Failed to update players list in Redis for room %s: %v", roomID, err)
		return nil, false, fmt.Errorf("failed to update players list in Redis: %w", err)
	}

	// 4. Удаляем специфичные для ушедшего игрока поля из Redis
	if err := s.roomStateRepo.DeletePlayerSpecificFields(ctx, roomID, leavingUserID, playerSpecificBaseFields); err != nil {
		log.Printf("Use Case LeaveRoom: Warning - failed to delete specific fields for player %s in room %s: %v", leavingUserID, roomID, err)
	}

	// 5. Если после ухода игрока комната стала пустой, удаляем ее полностью из Redis.
	if len(remainingPlayerIDsAfterLeave) == 0 {
		log.Printf("Use Case LeaveRoom: Room %s is now empty. Deleting room from Redis.", roomID)
		if delErr := s.roomStateRepo.DeleteRoom(ctx, roomID); delErr != nil {
			log.Printf("Use Case LeaveRoom: Failed to delete empty room %s from Redis: %v", roomID, delErr)
			return nil, false, fmt.Errorf("room is empty but failed to delete from redis: %w", delErr) // Более строгая обработка
		}
		log.Printf("Use Case LeaveRoom: Player %s left room %s, room is now empty and was successfully deleted.", leavingUserID, roomID)
		return nil, true, nil
	}

	wasRoomDeleted = false

	var remainingPlayerSingleID string
	if len(remainingPlayerIDsAfterLeave) == 1 {
		remainingPlayerSingleID = remainingPlayerIDsAfterLeave[0]
		log.Printf("Use Case LeaveRoom: One player %s remains in room %s. Resetting their state and room status.", remainingPlayerSingleID, roomID)

		if err := s.roomStateRepo.ResetPlayerState(ctx, roomID, remainingPlayerSingleID); err != nil {
			log.Printf("Use Case LeaveRoom: Failed to reset state for remaining player %s in room %s: %v", remainingPlayerSingleID, roomID, err)
		}
		if err := s.roomStateRepo.SetRoomField(ctx, roomID, "status", "waiting"); err != nil {
			log.Printf("Use Case LeaveRoom: Failed to set room status to 'waiting' for room %s: %v", roomID, err)
		}
		if err := s.roomStateRepo.SetRoomField(ctx, roomID, "turn", ""); err != nil {
			log.Printf("Use Case LeaveRoom: Failed to reset turn for room %s: %v", roomID, err)
		}
	}

	// 7. Конструируем обновленную модель model.Room с оставшимися игроками
	finalPlayersInModel := make([]*model.Player, 0, len(remainingPlayerIDsAfterLeave))
	for _, pID := range remainingPlayerIDsAfterLeave {
		player := &model.Player{ID: pID}
		if len(remainingPlayerIDsAfterLeave) == 1 && pID == remainingPlayerSingleID {
			player.IsReady = false
			player.Score = 0
			player.LastAction = ""
			player.Hand = []model.Card{}
		}

		finalPlayersInModel = append(finalPlayersInModel, player)
	}

	roomBet, _ := strconv.Atoi(roomStateMap["bet"])

	currentStatus := roomStateMap["status"]
	currentTurn := roomStateMap["turn"]
	if len(remainingPlayerIDsAfterLeave) == 1 {
		currentStatus = "waiting"
		currentTurn = ""
	}

	resultRoomModel := &model.Room{
		ID:                  roomID,
		Status:              currentStatus,
		Bet:                 roomBet,
		Players:             finalPlayersInModel,
		CurrentTurnPlayerID: currentTurn,
		Deck:                []model.Card{},
	}

	log.Printf("Use Case LeaveRoom: Player %s left room %s. Remaining players: %d. Status: %s",
		leavingUserID, roomID, len(resultRoomModel.Players), resultRoomModel.Status)
	return resultRoomModel, wasRoomDeleted, nil
}

// parseCardString преобразует строку типа "AS" в model.Card{Value:"A", Suit:"S"}
func parseCardString(cardStr string) (model.Card, bool) {
	if len(cardStr) < 2 { // Карта должна иметь хотя бы значение и масть
		return model.Card{}, false
	}

	value := ""
	suit := ""

	// Простой парсер: последняя буква - масть, все остальное - значение
	// Например, "10H" -> Value "10", Suit "H"
	// "AS" -> Value "A", Suit "S"
	suit = string(cardStr[len(cardStr)-1])
	value = cardStr[:len(cardStr)-1]

	// Простая валидация масти (можно расширить)
	validSuits := map[string]bool{"H": true, "D": true, "C": true, "S": true}
	if !validSuits[suit] {
		return model.Card{}, false
	}

	// Простая валидация значения (можно расширить)
	// ... (здесь можно добавить проверку на A, K, Q, J, 2-10)

	return model.Card{Value: value, Suit: suit}, true
}

// parseHandString преобразует строку из Redis (например, "AS,KD,10H" или "nil") в срез model.Card
func parseHandString(handStr string) []model.Card {
	if handStr == "nil" || handStr == "" {
		return []model.Card{}
	}

	cardStrings := strings.Split(handStr, ",")
	hand := make([]model.Card, 0, len(cardStrings))

	for _, cs := range cardStrings {
		trimmedCS := strings.TrimSpace(cs)
		if card, ok := parseCardString(trimmedCS); ok {
			hand = append(hand, card)
		}
	}
	return hand
}
