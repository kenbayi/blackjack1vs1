package handlers

import (
	"blackjack/src/db"
	"log"
	"strings"
)

func (h *Hub) leaveRoom(msg Message) {
	content := msg.Content.(map[string]interface{})
	roomID := content["roomID"].(string)
	playerID := content["playerID"].(string)

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

	// Fetch room state from Redis hash
	roomState, err := db.RedisClient.HGetAll(db.Ctx, "room:"+roomID).Result()
	if err != nil {
		msg.Conn.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": "Error retrieving room state",
		})
		return
	}

	// Check if the player exists in the room
	players := roomState["players"]
	if !strings.Contains(players, playerID) {
		msg.Conn.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": "Player not in room",
		})
		return
	}

	// Remove the player from the players list in Redis (comma-separated string)
	playersList := strings.Split(players, ",")
	var updatedPlayersList []string
	for _, p := range playersList {
		if p != playerID {
			updatedPlayersList = append(updatedPlayersList, p)
		}
	}
	updatedPlayers := strings.Join(updatedPlayersList, ",")
	err = db.RedisClient.HSet(db.Ctx, "room:"+roomID, "players", updatedPlayers).Err()
	if err != nil {
		log.Println("Error updating players list:", err)
		return
	}

	// Remove player-specific data from Redis
	// Remove ready status, hand, last action, and score for the player
	err = db.RedisClient.HDel(db.Ctx, "room:"+roomID, "readyStatus."+playerID).Err()
	if err != nil {
		log.Println("Error deleting ready status:", err)
	}

	err = db.RedisClient.HDel(db.Ctx, "room:"+roomID, "hands."+playerID).Err()
	if err != nil {
		log.Println("Error deleting hand:", err)
	}

	err = db.RedisClient.HDel(db.Ctx, "room:"+roomID, "lastAction."+playerID).Err()
	if err != nil {
		log.Println("Error deleting last action:", err)
	}

	err = db.RedisClient.HDel(db.Ctx, "room:"+roomID, "scores."+playerID).Err()
	if err != nil {
		log.Println("Error deleting score:", err)
	}

	// Remove player from in-memory room
	h.mu.Lock()
	delete(room.Players, msg.Conn)
	h.mu.Unlock()

	// Notify all clients about the updated room state
	h.broadcastRoom(roomID, map[string]interface{}{
		"type": "room_left",
		"players": func() []string {
			players := []string{}
			for _, id := range room.Players {
				players = append(players, id)
			}
			return players
		}(),
	})

	// Notify all clients about the updated player list
	h.broadcastAll(map[string]interface{}{
		"type":   "update_list",
		"action": "leave",
		"roomID": roomID,
		"players": func() []string {
			players := []string{}
			for _, id := range room.Players {
				players = append(players, id)
			}
			return players
		}(),
	})

	// Optionally, if the room is now empty, you can delete the room data from Redis
	if len(updatedPlayers) == 0 {
		err = db.RedisClient.Del(db.Ctx, "room:"+roomID).Err()
		if err != nil {
			log.Println("Error deleting room from Redis:", err)
		}
	}
}
