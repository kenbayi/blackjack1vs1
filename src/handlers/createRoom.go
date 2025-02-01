package handlers

import (
	"blackjack/src/db"
	"blackjack/src/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
)

func (h *Hub) createRoom(msg Message) {
	content := msg.Content.(map[string]interface{})
	playerID := content["playerID"].(string)
	bet := int(content["bet"].(float64))

	// Validate if the player has enough money
	playerMoney, err := models.GetPlayerBalance(db.PostgresDB, playerID)
	if err != nil || playerMoney < bet {
		msg.Conn.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": "Insufficient funds to create a room",
		})
		return
	}

	// Generate a unique room ID
	roomID := generateRoomID()

	// Initialize Redis hash fields
	roomState := map[string]interface{}{
		"roomID":                  roomID,
		"status":                  "waiting",
		"bet":                     bet,
		"players":                 playerID, // single player ID for now
		"readyStatus." + playerID: false,    // default ready status for the creator
		"hands." + playerID:       nil,      // empty hand for the creator
		"lastAction." + playerID:  nil,      // no action yet
		"scores." + playerID:      0,        // initial score for the creator
		"deck":                    nil,
	}

	// Save the room state directly in the Redis hash
	err = db.RedisClient.HMSet(db.Ctx, "room:"+roomID, roomState).Err()
	if err != nil {
		log.Println("Error saving room state to Redis:", err)
		return
	}

	// Create Room in memory
	room := &Room{
		ID:      roomID,
		Players: map[*websocket.Conn]string{msg.Conn: playerID},
	}

	h.mu.Lock()
	h.Rooms[roomID] = room
	h.mu.Unlock()

	// Notify creator about the room creation
	msg.Conn.WriteJSON(map[string]interface{}{
		"type":   "room_created",
		"roomID": roomID,
	})

	// Broadcast room creation
	h.broadcastAll(map[string]interface{}{
		"type":    "update_list",
		"action":  "create",
		"roomID":  roomID,
		"status":  "waiting",
		"players": []string{playerID},
		"bet":     bet,
	})
}

func generateRoomID() string {
	// Placeholder function to generate unique room IDs
	return uuid.New().String()
}
