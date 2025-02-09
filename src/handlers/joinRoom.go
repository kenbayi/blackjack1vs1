package handlers

import (
	"blackjack/src/db"
	"blackjack/src/models"
	"log"
)

func (h *Hub) joinRoom(msg Message) {
	content := msg.Content.(map[string]interface{})
	roomID := content["roomID"].(string)
	playerID := content["playerID"].(string)
	bet := int(content["bet"].(float64))

	h.mu.Lock()
	room, exists := h.Rooms[roomID]
	h.mu.Unlock()

	if !exists {
		msg.Conn.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": "Room not found",
		})
		return
	}

	if len(room.Players) >= 2 {
		msg.Conn.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": "Room is full",
		})
		return
	}

	// Fetch room state from Redis hash
	roomState, err := db.RedisClient.HGetAll(db.Ctx, "room:"+roomID).Result()
	if err != nil {
		msg.Conn.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": "Error retrieving room state",
		})
		return
	}

	// Validate the joining player's money
	playerMoney, err := models.GetPlayerBalance(db.PostgresDB, playerID)
	if err != nil || playerMoney < bet {
		msg.Conn.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": "Insufficient funds to join the room",
		})
		return
	}

	// Update room state fields in Redis hash
	// Update players list (append the new player)
	players := roomState["players"]
	players += "," + playerID
	err = db.RedisClient.HSet(db.Ctx, "room:"+roomID, "players", players).Err()
	if err != nil {
		log.Println("Error updating players list:", err)
		return
	}

	// Set the ready status for the new player
	err = db.RedisClient.HSet(db.Ctx, "room:"+roomID, "readyStatus."+playerID, false).Err()
	if err != nil {
		log.Println("Error setting ready status for player:", err)
		return
	}

	// Set the player's hand as nil
	err = db.RedisClient.HSet(db.Ctx, "room:"+roomID, "hands."+playerID, "nil").Err()
	if err != nil {
		log.Println("Error setting hand for player:", err)
		return
	}

	// Set the player's last action as nil
	err = db.RedisClient.HSet(db.Ctx, "room:"+roomID, "lastAction."+playerID, "nil").Err()
	if err != nil {
		log.Println("Error setting last action for player:", err)
		return
	}

	// Set the player's score as 0
	err = db.RedisClient.HSet(db.Ctx, "room:"+roomID, "scores."+playerID, 0).Err()
	if err != nil {
		log.Println("Error setting score for player:", err)
		return
	}

	// Add player to in-memory room
	h.mu.Lock()
	room.Players[msg.Conn] = playerID
	h.mu.Unlock()

	h.broadcastRoom(roomID, map[string]interface{}{
		"type": "room_joined",
		"players": func() []string {
			players := []string{}
			for _, id := range room.Players {
				players = append(players, id)
			}
			return players
		}(),
	})

	// Notify players to press "Ready" for the next round
	h.broadcastRoom(roomID, map[string]interface{}{
		"type": "game_waiting",
		"msg":  "Both players need to press 'Ready' to start the next round.",
	})

	// Notify about the updated room state
	h.broadcastAll(map[string]interface{}{
		"type":   "update_list",
		"action": "join",
		"roomID": roomID,
		"players": func() []string {
			players := []string{}
			for _, id := range room.Players {
				players = append(players, id)
			}
			return players
		}(),
	})
}
