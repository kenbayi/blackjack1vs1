package handlers

import (
	"blackjack/src/db"
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

	removePlayer(players, playerID, roomID)

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

}
