package usecase

import (
	"context"
	"errors"
	"fmt"
	"game_svc/internal/adapter/ws/server/dto"
	"game_svc/pkg/def"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"game_svc/internal/model"
)

// GameServiceImpl реализует GameUseCase.
type GameServiceImpl struct {
	roomStateRepo   RoomStateRepository
	producer        GameEventStorage
	clientPresenter ClientPresenter
}

// NewGameService конструктор для GameServiceImpl.
func NewGameService(rsr RoomStateRepository, pr GameEventStorage, presenter ClientPresenter) *GameServiceImpl {
	return &GameServiceImpl{
		roomStateRepo:   rsr,
		producer:        pr,
		clientPresenter: presenter,
	}
}

func generateShuffledDeckForGame() []model.Card {
	suits := []string{"H", "D", "C", "S"}
	values := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	deck := make([]model.Card, 0, 4*52) // 4 колоды

	for i := 0; i < 4; i++ {
		for _, suit := range suits {
			for _, value := range values {
				deck = append(deck, model.Card{Value: value, Suit: suit})
			}
		}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	return deck
}

// dealCardFromDeck берет карту из переданной колоды (среза model.Card) и возвращает карту и обновленную колоду.
func dealCardFromDeck(deck []model.Card) (model.Card, []model.Card, bool) {
	if len(deck) == 0 {
		return model.Card{}, deck, false
	}
	card := deck[0]
	newDeck := deck[1:]
	return card, newDeck, true
}

func calculateScoreForHand(hand []model.Card) int {
	score := 0
	aces := 0
	for _, card := range hand {
		value := card.Value
		switch value {
		case "A":
			aces++
			score += 11
		case "K", "Q", "J":
			score += 10
		default:
			num, err := strconv.Atoi(value)
			if err == nil {
				score += num
			}
		}
	}
	for score > 21 && aces > 0 {
		score -= 10
		aces--
	}
	return score
}

func (s *GameServiceImpl) PlayerReady(params model.PlayerReadyParams) (*model.PlayerReadyResult, error) {
	ctx := context.Background()
	userID := params.UserID
	roomID := params.RoomID
	isReady := params.IsReady

	log.Printf("Use Case PlayerReady: User %s in room %s set ready to %t", userID, roomID, isReady)

	// 1. Обновляем статус готовности игрока в Redis
	readyValue := "0"
	if !isReady {
		readyValue = "1"
	}
	if err := s.roomStateRepo.SetRoomField(ctx, roomID, fmt.Sprintf("readyStatus.%s", userID), readyValue); err != nil {
		log.Printf("Use Case PlayerReady: Failed to set readyStatus for player %s in room %s: %v", userID, roomID, err)
		return nil, fmt.Errorf("failed to update player ready status: %w", err)
	}

	// 2. Получаем текущее состояние комнаты, чтобы проверить готовность всех и собрать модель
	roomStateMap, err := s.roomStateRepo.GetAllRoomFields(ctx, roomID)
	if err != nil || len(roomStateMap) == 0 {
		log.Printf("Use Case PlayerReady: Error retrieving room state for room %s or room not found: %v", roomID, err)
		return nil, fmt.Errorf("room not found or error retrieving state: %w", err)
	}

	playerIDsStr := roomStateMap["players"]
	allPlayerIDsInRoom := splitPlayers(playerIDsStr)

	if len(allPlayerIDsInRoom) == 0 {
		return nil, errors.New("no players found in room, cannot process ready status")
	}
	if len(allPlayerIDsInRoom) < 2 && isReady {
		log.Printf("Use Case PlayerReady: Not enough players in room %s to start game.", roomID)
		currentRoomModel := s.reconstructRoomModel(roomID, roomStateMap, allPlayerIDsInRoom, nil)
		return &model.PlayerReadyResult{
			UpdatedRoom:      currentRoomModel,
			GameJustStarted:  false,
			PlayerIDReady:    userID,
			IsPlayerNowReady: isReady,
		}, nil
	}

	// 3. Проверяем, все ли игроки готовы (логика из allReady)
	areAllPlayersReady := true
	if len(allPlayerIDsInRoom) < 2 {
		areAllPlayersReady = false
	} else {
		for _, pID := range allPlayerIDsInRoom {
			var currentReadyStr string
			if pID == userID {
				currentReadyStr = readyValue
			} else {
				currentReadyStr = roomStateMap[fmt.Sprintf("readyStatus.%s", pID)]
			}
			if currentReadyStr != "1" {
				areAllPlayersReady = false
				break
			}
		}
	}

	result := &model.PlayerReadyResult{
		PlayerIDReady:    userID,
		IsPlayerNowReady: isReady,
		GameJustStarted:  false,
	}

	if areAllPlayersReady && len(allPlayerIDsInRoom) == 2 {
		log.Printf("Use Case PlayerReady: All %d players ready in room %s. Starting game.", len(allPlayerIDsInRoom), roomID)
		result.GameJustStarted = true
		result.RoomRemovedFromList = true

		// --- Логика startGame ---
		// Обновляем статус комнаты в Redis
		if err := s.roomStateRepo.SetRoomField(ctx, roomID, "status", "in_progress"); err != nil {
			log.Printf("Use Case PlayerReady: Failed to set room status to 'in_progress' for room %s: %v", roomID, err)
			return nil, fmt.Errorf("failed to set room status: %w", err)
		}
		roomStateMap["status"] = "in_progress"

		turnPlayerID := allPlayerIDsInRoom[0]
		if err := s.roomStateRepo.SetRoomField(ctx, roomID, "turn", turnPlayerID); err != nil {
			log.Printf("Use Case PlayerReady: Failed to set turn for player %s in room %s: %v", turnPlayerID, roomID, err)
			return nil, fmt.Errorf("failed to set turn: %w", err)
		}
		roomStateMap["turn"] = turnPlayerID

		// Генерируем и перемешиваем колоду
		gameDeck := generateShuffledDeckForGame()

		// Раздаем по 2 карты каждому игроку
		playerHands := make(map[string][]model.Card)
		playerScores := make(map[string]int)

		for _, pID := range allPlayerIDsInRoom {
			var hand []model.Card
			var card1, card2 model.Card
			var deckAfterDeal1, deckAfterDeal2 []model.Card
			var ok bool

			card1, deckAfterDeal1, ok = dealCardFromDeck(gameDeck)
			if !ok {
				return nil, errors.New("deck ran out of cards during initial deal")
			}
			gameDeck = deckAfterDeal1

			card2, deckAfterDeal2, ok = dealCardFromDeck(gameDeck)
			if !ok {
				return nil, errors.New("deck ran out of cards during initial deal")
			}
			gameDeck = deckAfterDeal2

			hand = []model.Card{card1, card2}
			playerHands[pID] = hand
			playerScores[pID] = calculateScoreForHand(hand)

			// Обновляем руки и очки в Redis
			handStr := serializeHand(hand)
			if err := s.roomStateRepo.SetRoomField(ctx, roomID, fmt.Sprintf("hands.%s", pID), handStr); err != nil {
				return nil, fmt.Errorf("failed to set hand for player %s: %w", pID, err)
			}
			if err := s.roomStateRepo.SetRoomField(ctx, roomID, fmt.Sprintf("scores.%s", pID), strconv.Itoa(playerScores[pID])); err != nil {
				return nil, fmt.Errorf("failed to set score for player %s: %w", pID, err)
			}
			// Обновляем roomStateMap для конструирования модели
			roomStateMap[fmt.Sprintf("hands.%s", pID)] = handStr
			roomStateMap[fmt.Sprintf("scores.%s", pID)] = strconv.Itoa(playerScores[pID])
		}
		serializedDeck := serializeDeck(gameDeck)
		err := s.roomStateRepo.SetRoomField(ctx, roomID, "deck", serializedDeck)
		if err != nil {
			log.Printf("Use Case LeaveRoom: Failed to set room status to 'waiting' for room %s: %v", roomID, err)
		}
		roomStateMap["deck"] = serializedDeck

		result.UpdatedRoom = s.reconstructRoomModel(roomID, roomStateMap, allPlayerIDsInRoom, &gameDeck)
		result.UpdatedRoom.Deck = gameDeck
	} else {
		log.Printf("Use Case PlayerReady: Not all players ready in room %s, or not enough players.", roomID)
		result.UpdatedRoom = s.reconstructRoomModel(roomID, roomStateMap, allPlayerIDsInRoom, nil) // deck nil, т.к. игра не началась
	}

	return result, nil
}

func (s *GameServiceImpl) reconstructRoomModel(roomID string, roomStateMap map[string]string, playerIDs []string, deckToUse *[]model.Card) *model.Room {
	playersInModel := make([]*model.Player, 0, len(playerIDs))
	for _, pID := range playerIDs {
		player := &model.Player{ID: pID}

		readyStr := roomStateMap[fmt.Sprintf("readyStatus.%s", pID)] // Обновленное значение для текущего юзера уже должно быть в roomStateMap если мы его обновили
		player.IsReady = (readyStr == "1")

		scoreStr := roomStateMap[fmt.Sprintf("scores.%s", pID)]
		score, _ := strconv.Atoi(scoreStr)
		player.Score = score

		handStr := roomStateMap[fmt.Sprintf("hands.%s", pID)]
		player.Hand = parseHandString(handStr)

		player.LastAction = roomStateMap[fmt.Sprintf("lastAction.%s", pID)]

		playersInModel = append(playersInModel, player)
	}

	bet, _ := strconv.Atoi(roomStateMap["bet"])

	roomModel := &model.Room{
		ID:                  roomID,
		Status:              roomStateMap["status"],
		Bet:                 bet,
		Players:             playersInModel,
		CurrentTurnPlayerID: roomStateMap["turn"],
		Deck:                []model.Card{},
	}
	if deckToUse != nil {
		roomModel.Deck = *deckToUse
	}

	return roomModel
}

func serializeHand(hand []model.Card) string {
	if len(hand) == 0 {
		return "nil"
	}
	cardStrings := make([]string, len(hand))
	for i, card := range hand {
		cardStrings[i] = card.Value + card.Suit
	}
	return strings.Join(cardStrings, ",")
}

// serializeDeck converts a slice of model.Card (the deck) into a comma-separated string.
func serializeDeck(deck []model.Card) string {
	if len(deck) == 0 {
		return ""
	}
	cardStrings := make([]string, len(deck))
	for i, card := range deck {
		cardStrings[i] = card.Value + card.Suit
	}
	return strings.Join(cardStrings, ",")
}

func (s *GameServiceImpl) _endGameProcessing(ctx context.Context, roomID string, winnerID, loserID string, bet int, allPlayerIDs []string, finalHands map[string][]model.Card) error {
	log.Printf("Use Case: _endGameProcessing started for room %s. Winner: %s, Loser: %s, Bet: %d", roomID, winnerID, loserID, bet)

	// 1. Обновляем балансы игроков (если есть победитель и проигравший, и ставка > 0)
	if winnerID != "" && winnerID != "0" && loserID != "" && loserID != "0" && bet > 0 {
		winnerIDint, err := strconv.ParseInt(winnerID, 10, 64)
		loserIDint, err := strconv.ParseInt(loserID, 10, 64)
		if err != nil {
			return fmt.Errorf("use Case: Failed to parse winnerID: %w", err)
		}
		_, err = s.clientPresenter.AddBalance(ctx, model.User{ID: winnerIDint, Balance: def.Pointer(int64(bet))})
		if err != nil {
			return err
		}
		_, err = s.clientPresenter.SubtractBalance(ctx, model.User{ID: loserIDint, Balance: def.Pointer(int64(bet))})
		if err != nil {
			return err
		}
		log.Printf("Use Case: Winner %s gets %d, Loser %s loses %d", winnerID, bet, loserID, bet)
	}

	finalHandsStrForHistory := make(map[string][]string)
	if finalHands != nil {
		for pID, hand := range finalHands {
			handS := make([]string, len(hand))
			for i, card := range hand {
				handS[i] = card.Value + card.Suit
			}
			finalHandsStrForHistory[pID] = handS
		}
	}

	// 3. Сбрасываем состояние комнаты и игроков в Redis для новой игры
	log.Printf("Use Case _endGameProcessing: Resetting Redis state for room %s.", roomID)

	// Сбрасываем общие поля комнаты
	if err := s.roomStateRepo.SetRoomField(ctx, roomID, "status", "waiting"); err != nil {
		log.Printf("Use Case _endGameProcessing: Error resetting room status for room %s: %v", roomID, err)
		// Можно рассмотреть возврат ошибки, если это критично
	}
	if err := s.roomStateRepo.SetRoomField(ctx, roomID, "turn", ""); err != nil {
		log.Printf("Use Case _endGameProcessing: Error resetting turn for room %s: %v", roomID, err)
	}
	if err := s.roomStateRepo.SetRoomField(ctx, roomID, "deck", ""); err != nil { // Очищаем колоду
		log.Printf("Use Case _endGameProcessing: Error clearing deck for room %s: %v", roomID, err)
	}

	// Сбрасываем состояние каждого игрока, используя существующий метод репозитория
	for _, pID := range allPlayerIDs {
		if err := s.roomStateRepo.ResetPlayerState(ctx, roomID, pID); err != nil {
			log.Printf("Use Case _endGameProcessing: Error resetting state for player %s in room %s: %v", pID, roomID, err)
		}
	}

	log.Printf("Use Case _endGameProcessing: Room %s state fully reset for new game.", roomID)
	return nil
}

func (s *GameServiceImpl) Hit(params model.HitParams) (*model.Result, error) {
	ctx := context.Background()
	userID := params.UserID
	roomID := params.RoomID
	log.Printf("Use Case Hit: User %s in room %s", userID, roomID)

	result := &model.Result{RoomID: roomID, PlayerID: userID}

	roomStateMap, err := s.roomStateRepo.GetAllRoomFields(ctx, roomID)
	if err != nil || len(roomStateMap) == 0 {
		return nil, fmt.Errorf("room %s not found or error retrieving state: %w", roomID, err)
	}

	if roomStateMap["status"] != "in_progress" {
		return nil, errors.New("game is not in progress")
	}
	if roomStateMap["turn"] != userID {
		return nil, errors.New("not your turn")
	}

	deckStr := roomStateMap["deck"]
	currentDeck := parseHandString(deckStr)
	if len(currentDeck) == 0 {
		return nil, errors.New("deck is empty, cannot hit")
	}

	dealtCard, updatedDeck, dealtOK := dealCardFromDeck(currentDeck)
	if !dealtOK {
		return nil, errors.New("failed to deal card from deck")
	}
	result.DealtCard = &dealtCard

	playerHandStr := roomStateMap[fmt.Sprintf("hands.%s", userID)]
	playerHand := parseHandString(playerHandStr)
	playerHand = append(playerHand, dealtCard)
	result.PlayerHand = &playerHand
	newScore := calculateScoreForHand(playerHand)
	result.NewScore = &newScore

	allPlayerIDs := splitPlayers(roomStateMap["players"])
	opponentID := ""
	if len(allPlayerIDs) == 2 {
		if allPlayerIDs[0] == userID {
			opponentID = allPlayerIDs[1]
		} else {
			opponentID = allPlayerIDs[0]
		}
	} else {
		return nil, errors.New("invalid number of players for hit action")
	}

	if newScore > 21 {
		result.IsBusted = true
		result.GameEnded = true
		result.Winner = opponentID
		result.Loser = userID

		// Собираем FinalScores и FinalHands
		result.FinalScores = make(map[string]int)
		result.FinalHands = make(map[string][]model.Card)
		result.FinalScores[userID] = newScore
		result.FinalHands[userID] = playerHand

		opponentHand := parseHandString(roomStateMap[fmt.Sprintf("hands.%s", opponentID)])
		opponentScore, _ := strconv.Atoi(roomStateMap[fmt.Sprintf("scores.%s", opponentID)])
		result.FinalScores[opponentID] = opponentScore
		result.FinalHands[opponentID] = opponentHand

		roomBet, _ := strconv.Atoi(roomStateMap["bet"])
		errEnd := s._endGameProcessing(ctx, roomID, result.Winner, result.Loser, roomBet, allPlayerIDs, result.FinalHands)
		if errEnd != nil {
			log.Printf("Use Case Hit: Error during _endGameProcessing for room %s after bust: %v", roomID, errEnd)
		} else {
			err := s.producer.PushGameEnd(ctx, result, int64(roomBet))
			if err != nil {
				return nil, err
			}
		}
	} else {
		result.IsBusted = false
		result.GameEnded = false
		result.NextTurnPlayerID = opponentID
	}

	// Сохранение изменений в Redis
	pipeCmds := make(map[string]interface{})
	pipeCmds[fmt.Sprintf("hands.%s", userID)] = serializeHand(playerHand)
	pipeCmds[fmt.Sprintf("scores.%s", userID)] = strconv.Itoa(newScore)
	pipeCmds["deck"] = serializeDeck(updatedDeck)
	if !result.GameEnded {
		pipeCmds["turn"] = result.NextTurnPlayerID
	}

	for field, value := range pipeCmds {
		if errSet := s.roomStateRepo.SetRoomField(ctx, roomID, field, value); errSet != nil {
			log.Printf("Use Case Hit: Failed to set Redis field %s for room %s: %v", field, roomID, errSet)
			// Рассмотреть ошибку. Возможно, вернуть ее, если критично.
		}
	}
	return result, nil
}

func (s *GameServiceImpl) Stand(params model.StandParams) (*model.Result, error) {
	ctx := context.Background()
	userID := params.UserID
	roomID := params.RoomID
	log.Printf("Use Case Stand: User %s in room %s", userID, roomID)

	result := &model.Result{RoomID: roomID, PlayerID: userID}

	roomStateMap, err := s.roomStateRepo.GetAllRoomFields(ctx, roomID)
	if err != nil || len(roomStateMap) == 0 {
		return nil, fmt.Errorf("room %s not found or error retrieving state: %w", roomID, err)
	}

	if roomStateMap["status"] != "in_progress" {
		return nil, errors.New("game is not in progress")
	}
	if roomStateMap["turn"] != userID {
		return nil, errors.New("not your turn")
	}

	if err := s.roomStateRepo.SetRoomField(ctx, roomID, fmt.Sprintf("stood.%s", userID), "1"); err != nil {
		return nil, fmt.Errorf("failed to set stood status for player %s: %w", userID, err)
	}
	roomStateMap[fmt.Sprintf("stood.%s", userID)] = "1"

	allPlayerIDs := splitPlayers(roomStateMap["players"])
	opponentID := ""
	player1ID, player2ID := "", ""
	if len(allPlayerIDs) == 2 {
		player1ID = allPlayerIDs[0]
		player2ID = allPlayerIDs[1]
		if player1ID == userID {
			opponentID = player2ID
		} else {
			opponentID = player1ID
		}
	} else {
		return nil, errors.New("invalid number of players for stand action")
	}

	scoreUser, _ := strconv.Atoi(roomStateMap[fmt.Sprintf("scores.%s", userID)])
	scoreOpponent, _ := strconv.Atoi(roomStateMap[fmt.Sprintf("scores.%s", opponentID)])
	result.PlayerCurrentScore = &scoreUser
	result.AllPlayerScores = &map[string]int{userID: scoreUser, opponentID: scoreOpponent}

	opponentStood := roomStateMap[fmt.Sprintf("stood.%s", opponentID)] == "1"
	opponentBusted := false
	if scoreOpponent > 21 {
		opponentBusted = true
	}

	if opponentStood || opponentBusted {
		result.GameEnded = true
		log.Printf("Use Case Stand: Game ending condition met in room %s. Player %s stood. Opponent stood: %t, Opponent busted: %t", roomID, userID, opponentStood, opponentBusted)

		// Собираем руки
		playerHand := parseHandString(roomStateMap[fmt.Sprintf("hands.%s", userID)])
		opponentHand := parseHandString(roomStateMap[fmt.Sprintf("hands.%s", opponentID)])
		result.FinalHands = map[string][]model.Card{userID: playerHand, opponentID: opponentHand}
		result.FinalScores = map[string]int{userID: scoreUser, opponentID: scoreOpponent}

		if scoreUser > 21 {
			result.Winner = opponentID
			result.Loser = userID
		} else if scoreOpponent > 21 {
			result.Winner = userID
			result.Loser = opponentID
		} else if scoreUser > scoreOpponent {
			result.Winner = userID
			result.Loser = opponentID
		} else if scoreOpponent > scoreUser {
			result.Winner = opponentID
			result.Loser = userID
		} else { // Ничья
			result.Winner = "0"
			result.Loser = "0"
		}

		roomBet, _ := strconv.Atoi(roomStateMap["bet"])
		errEnd := s._endGameProcessing(ctx, roomID, result.Winner, result.Loser, roomBet, allPlayerIDs, result.FinalHands)
		if errEnd != nil {
			log.Printf("Use Case Stand: Error during _endGameProcessing for room %s: %v", roomID, errEnd)
		} else {
			err := s.producer.PushGameEnd(ctx, result, int64(roomBet))
			if err != nil {
				return nil, err
			}
		}
	} else {
		result.GameEnded = false
		result.NextTurnPlayerID = opponentID
		if err := s.roomStateRepo.SetRoomField(ctx, roomID, "turn", opponentID); err != nil {
			log.Printf("Use Case Stand: Failed to set turn for room %s: %v", roomID, err)
		}
	}
	return result, nil
}

func (s *GameServiceImpl) HandlePlayerDisconnect(disconnectedUserID string, roomID string) (*dto.DisconnectResponse, error) {
	ctx := context.Background()
	log.Printf("Use Case HandlePlayerDisconnect: User %s disconnected from room %s", disconnectedUserID, roomID)

	response := &dto.DisconnectResponse{
		LeftPlayerID: disconnectedUserID,
		RoomID:       roomID,
	}

	roomStateMap, err := s.roomStateRepo.GetAllRoomFields(ctx, roomID)
	if err != nil {
		log.Printf("Use Case HandlePlayerDisconnect: Error retrieving room state for room %s: %v", roomID, err)
		return nil, fmt.Errorf("error retrieving room state: %w", err)
	}
	if len(roomStateMap) == 0 {
		log.Printf("Use Case HandlePlayerDisconnect: Room %s not found in Redis. Already deleted or never existed.", roomID)
		response.IsRoomDeleted = true
		response.RoomRemovedFromList = true // Если комнаты нет, ее точно нет в списке
		return response, nil                // Комнаты нет, делать нечего
	}

	currentPlayersStr := roomStateMap["players"]
	allPlayerIDsInRoom := splitPlayers(currentPlayersStr)

	// Проверяем, был ли игрок действительно в этой комнате (на случай гонки состояний)
	wasPlayerInRoom := false
	var remainingPlayerIDs []string
	for _, pID := range allPlayerIDsInRoom {
		if pID == disconnectedUserID {
			wasPlayerInRoom = true
		} else {
			remainingPlayerIDs = append(remainingPlayerIDs, pID)
		}
	}

	if !wasPlayerInRoom {
		log.Printf("Use Case HandlePlayerDisconnect: Player %s was not found in room %s's player list (%s). No action taken.", disconnectedUserID, roomID, currentPlayersStr)
		// Игрок уже был удален или его там не было. Возвращаем текущее состояние (комната не удалена этим вызовом).
		// response.IsRoomDeleted = false; // По умолчанию
		// response.RoomRemovedFromList = false; // По умолчанию
		return response, nil // Или вернуть ошибку "player not found in room"
	}

	// Обновляем список игроков в Redis, удаляя отключившегося
	updatedPlayersStrRedis := strings.Join(remainingPlayerIDs, ",")
	if err := s.roomStateRepo.UpdatePlayerList(ctx, roomID, updatedPlayersStrRedis); err != nil {
		log.Printf("Use Case HandlePlayerDisconnect: Failed to update players list in Redis for room %s: %v", roomID, err)
		return nil, fmt.Errorf("failed to update players list in Redis: %w", err)
	}

	// Удаляем специфичные для отключившегося игрока поля
	if err := s.roomStateRepo.DeletePlayerSpecificFields(ctx, roomID, disconnectedUserID, playerSpecificBaseFields); err != nil {
		log.Printf("Use Case HandlePlayerDisconnect: Warning - failed to delete specific fields for player %s in room %s: %v", disconnectedUserID, roomID, err)
	}

	// Логика для игры 1 на 1: если один игрок отключается, другой выигрывает, игра завершается.
	roomBet, _ := strconv.Atoi(roomStateMap["bet"])
	gameStatus := roomStateMap["status"]

	if len(allPlayerIDsInRoom) == 2 && gameStatus == "in_progress" { // Если было 2 игрока и игра шла
		response.GameEnded = true
		response.RoomRemovedFromList = true // После завершения игры комната обычно убирается из активных списков или сбрасывается

		opponentID := ""
		if remainingPlayerIDs[0] != "" { // Должен остаться один
			opponentID = remainingPlayerIDs[0]
		}

		if opponentID != "" {
			response.GameEndData = &dto.GameEndData{ // GameEndData из usecase
				RoomID:  roomID,
				Winner:  opponentID, // Оставшийся игрок выигрывает
				Loser:   disconnectedUserID,
				Scores:  map[string]int{},          // Очки могут быть неактуальны или их нужно собрать из roomStateMap
				Hands:   map[string][]model.Card{}, // Руки тоже
				Message: fmt.Sprintf("Player %s disconnected, player %s wins by default.", disconnectedUserID, opponentID),
			}
			// Заполняем Scores и Hands для GameEndData из roomStateMap, если это нужно для истории
			// Собираем руки и очки для _endGameProcessing и для GameEndData
			currentHands := make(map[string][]model.Card)
			currentScores := make(map[string]int)

			for _, pID := range allPlayerIDsInRoom {
				handStr := roomStateMap["hands."+pID]
				currentHands[pID] = parseHandString(handStr)
				scoreStr := roomStateMap["score."+pID]
				score, _ := strconv.Atoi(scoreStr)
				currentScores[pID] = score
			}
			response.GameEndData.Scores = currentScores
			response.GameEndData.Hands = currentHands

			// Вызываем _endGameProcessing для обновления балансов, истории и сброса комнаты
			result := model.Result{
				RoomID:      roomID,
				Winner:      opponentID,
				Loser:       disconnectedUserID,
				FinalHands:  currentHands,
				FinalScores: currentScores,
			}
			err := s._endGameProcessing(ctx, roomID, opponentID, disconnectedUserID, roomBet, allPlayerIDsInRoom, currentHands)
			if err != nil {
				log.Printf("Use Case HandlePlayerDisconnect: Error during _endGameProcessing for room %s: %v", roomID, err)
			} else {
				err := s.producer.PushGameEnd(ctx, &result, int64(roomBet))
				if err != nil {
					return nil, err
				}
			}
		} else {
			log.Printf("Use Case HandlePlayerDisconnect: Room %s had 2 players, but opponent ID not found after disconnect.", roomID)
			response.GameEndData = &dto.GameEndData{RoomID: roomID, Winner: "0", Message: "Game ended due to disconnect, no winner determined."} // Ничья или системная ошибка
			err := s._endGameProcessing(ctx, roomID, "0", "0", roomBet, allPlayerIDsInRoom, nil)
			if err != nil {
				log.Printf("Use Case HandlePlayerDisconnect: Error during _endGameProcessing for room %s with no winner: %v", roomID, err)
			} else {
				result := model.Result{
					RoomID:      roomID,
					Winner:      "0",
					Loser:       "0",
					FinalHands:  map[string][]model.Card{},
					FinalScores: map[string]int{},
				}
				err := s.producer.PushGameEnd(ctx, &result, int64(roomBet))
				if err != nil {
					return nil, err
				}
			}
		}
	} else if len(remainingPlayerIDs) == 0 { // Если комната стала пустой
		response.IsRoomDeleted = true
		response.RoomRemovedFromList = true
		log.Printf("Use Case HandlePlayerDisconnect: Room %s is now empty. Deleting from Redis.", roomID)
		if delErr := s.roomStateRepo.DeleteRoom(ctx, roomID); delErr != nil {
			log.Printf("Use Case HandlePlayerDisconnect: Failed to delete empty room %s from Redis: %v", roomID, delErr)
			response.IsRoomDeleted = false // Не удалось удалить, флаг сбрасываем
		}
	} else if len(remainingPlayerIDs) == 1 && gameStatus == "waiting" { // Если остался 1 игрок и игра не шла (ожидание)
		// Сбрасываем состояние оставшегося игрока (ready=false, score=0)
		// и статус комнаты на "waiting", если он был другим.
		remainingSingleID := remainingPlayerIDs[0]
		log.Printf("Use Case HandlePlayerDisconnect: One player %s remains in waiting room %s. Resetting their state.", remainingSingleID, roomID)
		if err := s.roomStateRepo.ResetPlayerState(ctx, roomID, remainingSingleID); err != nil {
			log.Printf("Use Case HandlePlayerDisconnect: Failed to reset state for remaining player %s: %v", remainingSingleID, err)
		}
		if roomStateMap["status"] != "waiting" {
			if err := s.roomStateRepo.SetRoomField(ctx, roomID, "status", "waiting"); err != nil {
				log.Printf("Use Case HandlePlayerDisconnect: Failed to set room status to waiting: %v", err)
			}
		}
		response.RoomRemovedFromList = true // Комната изменилась, список нужно обновить
	}

	// Заполняем RemainingPlayersInRoom в ответе
	if !response.IsRoomDeleted && len(remainingPlayerIDs) > 0 {
		currentRoomStateAfterUpdates, _ := s.roomStateRepo.GetAllRoomFields(ctx, roomID) // Перечитываем, чтобы получить актуальные сброшенные значения

		for _, pID := range remainingPlayerIDs {
			player := &model.Player{ID: pID}
			if currentRoomStateAfterUpdates != nil { // Если комната еще существует
				readyStr := currentRoomStateAfterUpdates["readyStatus."+pID]
				player.IsReady = (readyStr == "1")
				scoreStr := currentRoomStateAfterUpdates["score."+pID]
				score, _ := strconv.Atoi(scoreStr)
				player.Score = score
				player.Hand = parseHandString(currentRoomStateAfterUpdates["hands."+pID])
				player.LastAction = currentRoomStateAfterUpdates["lastAction."+pID]
			} else {
				break
			}
			response.RemainingPlayersInRoom = append(response.RemainingPlayersInRoom, player)
		}
	}

	log.Printf("Use Case HandlePlayerDisconnect: Processed for user %s in room %s. GameEnded: %t, RoomDeleted: %t",
		disconnectedUserID, roomID, response.GameEnded, response.IsRoomDeleted)
	return response, nil
}
