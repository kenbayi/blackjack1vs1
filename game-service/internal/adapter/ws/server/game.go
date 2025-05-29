package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"game_svc/internal/model"
	"log"

	"game_svc/internal/adapter/ws/server/dto"
	gameservicews "game_svc/pkg/ws"
)

type GameMessageHandler struct {
	roomUseCase   RoomUseCase
	gameUseCase   GameUseCase
	rankedUseCase RankedUseCase
	hub           *gameservicews.Hub
}

func NewGameMessageHandler(
	hub *gameservicews.Hub,
	roomUC RoomUseCase,
	gameUC GameUseCase,
	rankedUC RankedUseCase,
) *GameMessageHandler {
	if hub == nil || roomUC == nil || gameUC == nil || rankedUC == nil {
		log.Fatal("GameMessageHandler: Cannot create with nil dependencies")
	}
	return &GameMessageHandler{
		hub:           hub,
		roomUseCase:   roomUC,
		gameUseCase:   gameUC,
		rankedUseCase: rankedUC,
	}
}

func (gmh *GameMessageHandler) Handle(rawMsg *gameservicews.RawMessage) {
	client := rawMsg.Client
	log.Printf("GameMessageHandler: Received message from client %s (UserID: %s, RoomID: %s), Payload: %s",
		client.Conn.RemoteAddr(), client.UserID, client.RoomID, string(rawMsg.Payload))

	var msg dto.GameMessage
	if err := json.Unmarshal(rawMsg.Payload, &msg); err != nil {
		log.Printf("GameMessageHandler: Error unmarshalling message from client %s: %v", client.UserID, err)
		gmh.sendErrorToClient(client, "invalid_message_format", "Could not parse message.")
		return
	}

	if client.UserID == "" {
		log.Printf("GameMessageHandler: Denying message from unauthenticated client %s", client.Conn.RemoteAddr())
		gmh.sendErrorToClient(client, "authentication_required", "User ID is missing.")
		return
	}

	var err error
	switch msg.Type {
	case "create_room":
		err = gmh.handleCreateRoom(client, msg.Payload)
	case "join_room":
		err = gmh.handleJoinRoom(client, msg.Payload)
	case "leave_room":
		err = gmh.handleLeaveRoom(client)
	case "ready":
		err = gmh.handleReady(client, msg.Payload)
	case "hit":
		err = gmh.handleHit(client)
	case "stand":
		err = gmh.handleStand(client)
	case "find_ranked_match":
		err = gmh.handleFindRankedMatch(client)
	default:
		log.Printf("GameMessageHandler: Unknown message type '%s' from client %s", msg.Type, client.UserID)
		gmh.sendErrorToClient(client, "unknown_message_type", fmt.Sprintf("Unknown message type: %s", msg.Type))
		return
	}

	if err != nil {
		log.Printf("GameMessageHandler: Error handling message type '%s' for client %s: %v", msg.Type, client.UserID, err)
	}
}

func (gmh *GameMessageHandler) handleCreateRoom(client *gameservicews.Client, payload interface{}) error {
	var req dto.CreateRoomPayload
	if err := dto.MapToStruct(payload, &req); err != nil {
		gmh.sendErrorToClient(client, "invalid_payload", "Could not parse create_room payload.")
		return fmt.Errorf("parsing create_room payload: %w", err)
	}

	if req.Bet <= 0 {
		err := errors.New("bet must be a positive value")
		gmh.sendErrorToClient(client, "invalid_bet", err.Error())
		return err
	}

	ucParams := dto.FromCreateRequestToParams(req, client.UserID)

	ucResponse, err := gmh.roomUseCase.CreateRoom(*ucParams)
	if err != nil {
		gmh.sendErrorToClient(client, "create_room_failed", err.Error())
		return err
	}

	// После успешного создания комнаты в use case, присваиваем RoomID клиенту
	client.RoomID = ucResponse.ID

	notification := dto.FromModelToListResponse(ucResponse)

	// Оповещаем создателя
	gmh.sendToClient(client, "room_created", notification.RoomID)

	// Оповещаем всех об обновлении списка комнат
	gmh.broadcastAll("update_list", notification)

	log.Printf("User %s created room %s with bet %d", client.UserID, ucResponse.ID, req.Bet)
	return nil
}

func (gmh *GameMessageHandler) handleJoinRoom(client *gameservicews.Client, payload interface{}) error {
	var req dto.JoinRoomPayload
	if err := dto.MapToStruct(payload, &req); err != nil {
		gmh.sendErrorToClient(client, "invalid_payload", "Could not parse join_room payload.")
		return fmt.Errorf("parsing join_room payload: %w", err)
	}
	if req.RoomID == "" {
		err := errors.New("room_id cannot be empty")
		gmh.sendErrorToClient(client, "invalid_payload", err.Error())
		return err
	}

	ucParams := dto.FromJoinRequestToParams(req, client.UserID)

	ucResponse, err := gmh.roomUseCase.JoinRoom(*ucParams)
	if err != nil {
		gmh.sendErrorToClient(client, "join_room_failed", err.Error())
		return err
	}
	client.RoomID = req.RoomID

	notification := dto.FromModelToListResponse(ucResponse)

	gmh.broadcastToRoom(req.RoomID, "room_joined", notification.Players)
	gmh.broadcastAll("update_list", notification)
	gmh.broadcastToRoom(req.RoomID, "game_waiting", "Both players need to press 'Ready' to start the next round.")

	log.Printf("User %s joined room %s", client.UserID, req.RoomID)
	return nil
}

func (gmh *GameMessageHandler) handleLeaveRoom(client *gameservicews.Client) error {
	if client.RoomID == "" {
		gmh.sendErrorToClient(client, "not_in_room", "You are not currently in a room.")
		return nil
	}

	roomIDToLeave := client.RoomID
	userID := client.UserID
	log.Printf("Handler: User %s attempting to leave room %s", userID, roomIDToLeave)

	ucParams := dto.FromLeaveRequestToParams(roomIDToLeave, userID)

	updatedRoomModel, wasRoomDeleted, err := gmh.roomUseCase.LeaveRoom(*ucParams)

	if err != nil {
		gmh.sendErrorToClient(client, "leave_room_failed", err.Error())
		if client.RoomID == roomIDToLeave {
			client.RoomID = ""
		}
		return err
	}

	if client.RoomID == roomIDToLeave {
		client.RoomID = ""
	}

	// 1. Оповещаем самого клиента, что он успешно вышел
	gmh.sendToClient(client, "left_room_successfully", "you have left the room")

	// 2. Оповещаем оставшихся игроков в комнате (если комната не удалена и есть кто-то)
	if !wasRoomDeleted && updatedRoomModel != nil && len(updatedRoomModel.Players) > 0 {
		playerLeftNotification := dto.FromLeaveRequestToPlayerLeftNotification(updatedRoomModel, userID)
		gmh.broadcastToRoom(updatedRoomModel.ID, "room_left", playerLeftNotification)
	} else if !wasRoomDeleted && updatedRoomModel != nil && len(updatedRoomModel.Players) == 0 {
		log.Printf("Handler: Room %s became empty after player left, but not marked as deleted by use case.", updatedRoomModel.ID)
	}

	// 3. Оповещаем всех об обновлении списка комнат
	roomListUpdateData := dto.FromLeaveRequestToUpdateList(roomIDToLeave, updatedRoomModel, userID, wasRoomDeleted)
	gmh.broadcastAll("update_list", roomListUpdateData)

	log.Printf("Handler: User %s successfully left room %s. Use case reported room deleted: %t",
		userID, roomIDToLeave, wasRoomDeleted)
	return nil
}

func (gmh *GameMessageHandler) handleReady(client *gameservicews.Client, payload interface{}) error {
	if client.RoomID == "" {
		gmh.sendErrorToClient(client, "not_in_room", "You must be in a room to set ready status.")
		return errors.New("client not in a room for ready")
	}

	var reqPayload dto.ReadyPayload
	if err := dto.MapToStruct(payload, &reqPayload); err != nil {
		gmh.sendErrorToClient(client, "invalid_payload", "Could not parse ready payload.")
		return fmt.Errorf("parsing ready payload: %w", err)
	}

	ucParams := model.PlayerReadyParams{
		UserID:  client.UserID,
		RoomID:  client.RoomID,
		IsReady: reqPayload.IsReady,
	}

	ucResult, err := gmh.gameUseCase.PlayerReady(ucParams)
	if err != nil {
		gmh.sendErrorToClient(client, "set_ready_failed", err.Error())
		return err
	}

	if ucResult.UpdatedRoom == nil {
		log.Printf("Handler handleReady: Received nil UpdatedRoom from use case for room %s without error.", client.RoomID)
		gmh.sendErrorToClient(client, "internal_error", "Failed to process ready status.")
		return errors.New("use case returned nil room without error on ready")
	}

	// Оповещаем комнату о статусе готовности игрока или о старте игры
	if !ucResult.GameJustStarted {
		playerReadyMsg := map[string]interface{}{
			"playerReady": ucResult.PlayerIDReady,
		}
		gmh.broadcastToRoom(ucResult.UpdatedRoom.ID, "player_ready", playerReadyMsg)
		log.Printf("Handler: Player %s in room %s is now %s. Waiting for other players.",
			ucResult.PlayerIDReady, ucResult.UpdatedRoom.ID, map[bool]string{true: "ready", false: "not ready"}[ucResult.IsPlayerNowReady])

	} else {
		gameStartDTO := dto.FromRoomModelToGameStateUpdate(ucResult.UpdatedRoom, "Game started! Initial cards dealt.")
		if gameStartDTO == nil {
			log.Printf("Handler handleReady: Failed to map room model to game start DTO for room %s", ucResult.UpdatedRoom.ID)
			gmh.sendErrorToClient(client, "internal_error", "Failed to prepare game start message.")
			return errors.New("failed to map room model to game start DTO")
		}
		gmh.broadcastToRoom(ucResult.UpdatedRoom.ID, "game_started", gameStartDTO.State)
		log.Printf("Handler: Game started in room %s. Initial state sent.", ucResult.UpdatedRoom.ID)

		if ucResult.UpdatedRoom.Status == "in_progress" {
			updateListMsg := dto.RoomListUpdateDTO{
				Action: "remove",
				RoomID: ucResult.UpdatedRoom.ID,
			}
			gmh.broadcastAll("update_list", updateListMsg)
			log.Printf("Handler: Room %s removed from public list as game started.", ucResult.UpdatedRoom.ID)
		}
	}
	return nil
}

func (gmh *GameMessageHandler) handleHit(client *gameservicews.Client) error {
	if client.RoomID == "" {
		gmh.sendErrorToClient(client, "not_in_room", "You must be in a room to hit.")
		return errors.New("client not in a room for hit")
	}

	ucParams := model.HitParams{UserID: client.UserID, RoomID: client.RoomID}
	ucResult, err := gmh.gameUseCase.Hit(ucParams)

	if err != nil {
		if err.Error() == "not your turn" {
			gmh.sendToClient(client, "warning", map[string]interface{}{"roomID": client.RoomID, "msg": "Not your turn"})
		} else {
			gmh.sendErrorToClient(client, "hit_failed", err.Error())
		}
		return err
	}

	if ucResult == nil {
		log.Printf("Handler handleHit: Received nil HitResult from use case for room %s without error.", client.RoomID)
		gmh.sendErrorToClient(client, "internal_error", "Failed to process hit.")
		return errors.New("use case returned nil result without error on hit")
	}

	// 1. Broadcast "hit" event (как в твоем старом коде)
	gmh.broadcastToRoom(ucResult.RoomID, "hit", map[string]interface{}{
		"forPlayer": ucResult.PlayerID,
		"card":      cardToString(*ucResult.DealtCard), // Преобразуем model.Card в строку
		"score":     ucResult.NewScore,
	})

	// 2. Если игрок перебрал (busted)
	if ucResult.IsBusted {
		gmh.broadcastToRoom(ucResult.RoomID, "busted", map[string]interface{}{
			"forPlayer": ucResult.PlayerID,
			"msg":       "Player busted!",
		})
	}

	// 3. Если игра завершилась (например, из-за bust)
	if ucResult.GameEnded {
		// Формируем руки для ответа (map[string][]string)
		finalHandsStr := make(map[string][]string)
		if ucResult.FinalHands != nil {
			for playerID, hand := range ucResult.FinalHands {
				handS := make([]string, len(hand))
				for i, card := range hand {
					handS[i] = cardToString(card)
				}
				finalHandsStr[playerID] = handS
			}
		}

		gmh.broadcastToRoom(ucResult.RoomID, "game_end", map[string]interface{}{
			"roomID": ucResult.RoomID,
			"winner": ucResult.Winner,      // Будет ID оппонента или "0"
			"scores": ucResult.FinalScores, // map[string]int
			"hands":  finalHandsStr,        // map[string][]string
		})
		gmh.broadcastToRoom(ucResult.RoomID, "game_waiting", map[string]interface{}{
			"msg": "Both players need to press 'Ready' to start the next round.",
		})
		log.Printf("Handler: Game ended in room %s after HIT by %s. Winner: %s", ucResult.RoomID, ucResult.PlayerID, ucResult.Winner)
	} else if !ucResult.IsBusted { // Если не bust и игра не закончилась, передаем ход
		gmh.broadcastToRoom(ucResult.RoomID, "turn", map[string]interface{}{
			"turn": ucResult.NextTurnPlayerID,
		})
		log.Printf("Handler: Turn changed in room %s to %s after HIT by %s", ucResult.RoomID, ucResult.NextTurnPlayerID, ucResult.PlayerID)
	}
	return nil
}

func cardToString(card model.Card) string {
	return card.Value + card.Suit
}

func (gmh *GameMessageHandler) handleStand(client *gameservicews.Client) error {
	if client.RoomID == "" {
		gmh.sendErrorToClient(client, "not_in_room", "You are not currently in a room to stand.")
		return errors.New("client not in a room for stand")
	}

	ucParams := model.StandParams{UserID: client.UserID, RoomID: client.RoomID}
	ucResult, err := gmh.gameUseCase.Stand(ucParams) // ucResult это *usecase.StandResult

	if err != nil {
		if err.Error() == "not your turn" {
			gmh.sendToClient(client, "warning", map[string]interface{}{"roomID": client.RoomID, "msg": "Not your turn"})
		} else {
			gmh.sendErrorToClient(client, "stand_failed", err.Error())
		}
		return err
	}
	if ucResult == nil {
		log.Printf("Handler handleStand: Received nil StandResult from use case for room %s without error.", client.RoomID)
		gmh.sendErrorToClient(client, "internal_error", "Failed to process stand.")
		return errors.New("use case returned nil result without error on stand")
	}

	gmh.broadcastToRoom(ucResult.RoomID, "stand", map[string]interface{}{
		"forPlayer": ucResult.PlayerID,
		"scores":    ucResult.AllPlayerScores,
	})

	// 2. Если игра завершилась (например, оба "stand")
	if ucResult.GameEnded {
		finalHandsStr := make(map[string][]string)
		if ucResult.FinalHands != nil {
			for playerID, hand := range ucResult.FinalHands {
				handS := make([]string, len(hand))
				for i, card := range hand {
					handS[i] = cardToString(card)
				}
				finalHandsStr[playerID] = handS
			}
		}

		gmh.broadcastToRoom(ucResult.RoomID, "game_end", map[string]interface{}{
			"roomID": ucResult.RoomID,
			"winner": ucResult.Winner,
			"scores": ucResult.FinalScores,
			"hands":  finalHandsStr,
		})
		gmh.broadcastToRoom(ucResult.RoomID, "game_waiting", map[string]interface{}{
			"msg": "Both players need to press 'Ready' to start the next round.",
		})
		log.Printf("Handler: Game ended in room %s after STAND by %s. Winner: %s", ucResult.RoomID, ucResult.PlayerID, ucResult.Winner)
	} else {
		gmh.broadcastToRoom(ucResult.RoomID, "turn", map[string]interface{}{
			"turn": ucResult.NextTurnPlayerID,
		})
		log.Printf("Handler: Turn changed in room %s to %s after STAND by %s", ucResult.RoomID, ucResult.NextTurnPlayerID, ucResult.PlayerID)
	}
	return nil
}

func (gmh *GameMessageHandler) sendErrorToClient(client *gameservicews.Client, errorType string, message string) {
	errorResp := dto.ErrorResponse{
		ErrorType: errorType,
		Message:   message,
	}
	gmh.sendToClient(client, "error", errorResp)
}

func (gmh *GameMessageHandler) sendToClient(client *gameservicews.Client, messageType string, content interface{}) {
	response := gameservicews.OutboundMessage{
		Type:    messageType,
		Content: content,
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("GameMessageHandler: Error marshalling message for client %s (type: %s): %v", client.UserID, messageType, err)
		return
	}
	gmh.hub.BroadcastToClient(client, jsonResponse)
}

func (gmh *GameMessageHandler) broadcastToRoom(roomID string, messageType string, content interface{}) {
	if roomID == "" {
		log.Printf("GameMessageHandler: Attempt to broadcast to empty roomID (type: %s). Aborted.", messageType)
		return
	}
	response := gameservicews.OutboundMessage{
		Type:    messageType,
		Content: content,
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("GameMessageHandler: Error marshalling message for room %s broadcast (type: %s): %v", roomID, messageType, err)
		return
	}
	gmh.hub.BroadcastToRoom(roomID, jsonResponse)
}

func (gmh *GameMessageHandler) broadcastAll(messageType string, content interface{}) {
	response := gameservicews.OutboundMessage{
		Type:    messageType,
		Content: content,
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("GameMessageHandler: Error marshalling message for broadcastAll (type: %s): %v", messageType, err)
		return
	}
	gmh.hub.BroadcastToAll(jsonResponse)
}

// HandlePlayerDisconnect обрабатывает логику, когда игрок неожиданно отключается.
// Этот метод вызывается из app.go через коллбэк OnDisconnectHandler хаба.
func (gmh *GameMessageHandler) HandlePlayerDisconnect(userID string, roomID string) {
	if userID == "" {
		log.Printf("Handler (HandlePlayerDisconnect): UserID is empty. Cannot process disconnect.")
		return
	}

	log.Printf("Handler: Processing disconnect for player %s from room %s", userID, roomID)

	ucResult, err := gmh.gameUseCase.HandlePlayerDisconnect(userID, roomID)
	if err != nil {
		log.Printf("Handler: Error from GameUseCase.HandlePlayerDisconnect for user %s in room %s: %v", userID, roomID, err)
		return
	}

	if ucResult == nil {
		log.Printf("Handler: GameUseCase.HandlePlayerDisconnect returned nil for user %s, room %s. No further action from handler.", userID, roomID)
		return
	}

	// Если игра продолжается с обновленным состоянием (например, ход перешел)
	if !ucResult.GameEnded && ucResult.UpdatedGameState != nil {
		// Сообщение об обновлении состояния игры
		gameStateNotification := dto.GameStateUpdate{ // Это ваш API DTO
			RoomID:  roomID,
			State:   ucResult.UpdatedGameState, // ucResult.UpdatedGameState должен быть совместим с тем, что ожидает фронт
			Message: fmt.Sprintf("Player %s disconnected. Game updated.", ucResult.LeftPlayerID),
		}
		gmh.broadcastToRoom(roomID, "game_state_update", gameStateNotification)
	}

	// Если игра завершилась из-за дисконнекта
	if ucResult.GameEnded && ucResult.GameEndData != nil {
		// Преобразуем ucResult.GameEndData.FinalHands в map[string][]string для API DTO
		finalHandsForAPI := dto.MapModelHandsToStringHandsForAPI(ucResult.GameEndData.Hands.(map[string][]model.Card)) // Требуется приведение типа

		gameEndAPIDTO := dto.GameEndBroadcastPayloadDTO{ // Это ваш API DTO
			RoomID: ucResult.GameEndData.RoomID,
			Winner: ucResult.GameEndData.Winner,
			Scores: ucResult.GameEndData.Scores.(map[string]int), // Требуется приведение типа
			Hands:  finalHandsForAPI,
		}
		gmh.broadcastToRoom(roomID, "game_end", gameEndAPIDTO)

		// Сообщение о ожидании новой игры
		// (message из GameEndData или стандартное)
		waitingMsg := "A player disconnected. Ready up for a new round."
		if ucResult.GameEndData.Message != "" {
			waitingMsg = ucResult.GameEndData.Message
		}
		gmh.broadcastToRoom(roomID, "game_waiting", map[string]string{"msg": waitingMsg})
	} else if !ucResult.GameEnded && ucResult.UpdatedGameState == nil {
		// Если игра не закончилась, состояние не обновилось (или не было данных),
		// но игрок точно ушел, и это не покрыто выше.
		playerLeftNotification := dto.PlayerLeftNotificationDTO{ // Это ваш API DTO
			RoomID:  roomID,
			Players: dto.GetPlayerIDsFromModels(ucResult.RemainingPlayersInRoom), // getPlayerIDsFromModels - вспомогательная
			Message: fmt.Sprintf("Player %s has disconnected.", ucResult.LeftPlayerID),
		}
		gmh.broadcastToRoom(roomID, "player_left", playerLeftNotification) // Используем "player_left" как в LeaveRoom
	}

	// Обновление общего списка комнат
	if ucResult.RoomRemovedFromList {
		action := "leave" // По умолчанию
		if ucResult.IsRoomDeleted {
			action = "remove"
		}
		updateListDTO := dto.RoomListUpdateDTO{ // Это ваш API DTO
			Action:  action,
			RoomID:  roomID, // roomID, из которой игрок вышел
			Players: dto.GetPlayerIDsFromModels(ucResult.RemainingPlayersInRoom),
			// Status и Bet можно взять из ucResult, если они там есть, или опустить для remove
		}
		gmh.broadcastAll("update_list", updateListDTO)
	}
	log.Printf("Handler: Processed disconnect for player %s from room %s. Use case reported room deleted: %t",
		userID, roomID, ucResult.IsRoomDeleted)
}

func (gmh *GameMessageHandler) handleFindRankedMatch(client *gameservicews.Client) error {
	log.Printf("User %s is searching for a ranked match.", client.UserID)
	gmh.sendToClient(client, "ranked_search_started", "Searching for an opponent...")

	// Your use case finds the match and returns the IDs
	match, err := gmh.rankedUseCase.FindMatch(client.UserID)
	if err != nil {
		gmh.sendErrorToClient(client, "ranked_search_failed", "An error occurred.")
		return err
	}

	// If no match was found yet, do nothing.
	if match == nil {
		return nil
	}

	// A match was found! `match` contains { RoomID, Players: [searcherID, opponentID] }
	log.Printf("Handler: Match found, RoomID: %s, Players: %v", match.RoomID, match.Players)

	searcherClient := client
	var opponentID string
	if searcherClient.UserID == match.Players[0] {
		opponentID = match.Players[1]
	} else {
		opponentID = match.Players[0]
	}

	opponentClient, ok := gmh.hub.GetClientByUserID(opponentID)
	if !ok {
		log.Printf("CRITICAL: Opponent %s disconnected before match could be finalized.", opponentID)
		gmh.sendErrorToClient(searcherClient, "match_failed", "Your opponent disconnected before the game could start.")
		return errors.New("opponent disconnected during match finalization")
	}

	log.Printf("Successfully retrieved client objects for searcher %s and opponent %s", searcherClient.UserID, opponentClient.UserID)

	searcherClient.RoomID = match.RoomID
	opponentClient.RoomID = match.RoomID

	log.Printf("In-memory state updated: Client %s -> Room %s, Client %s -> Room %s",
		searcherClient.UserID, searcherClient.RoomID, opponentClient.UserID, opponentClient.RoomID)

	notification := map[string]interface{}{
		"roomId": match.RoomID,
	}
	gmh.broadcastToRoom(match.RoomID, "match_found", notification)

	return nil
}
