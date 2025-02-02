package handlers

import (
	"blackjack/src/db"
	"log"
	"strings"
)

func (h *Hub) hitCard(msg Message) {
	content := msg.Content.(map[string]interface{})
	playerID := content["playerID"].(string)
	roomID := content["roomID"].(string)

	// Check if it's the player's turn
	turn := db.RedisClient.HGet(db.Ctx, "room:"+roomID, "turn").Val()
	if turn != playerID {
		log.Printf("Error: not your turn")

		// Message for sender about the turn
		msg.Conn.WriteJSON(map[string]interface{}{
			"type":   "warning",
			"roomID": roomID,
			"msg":    "Not your turn",
		})
		return
	}

	// Check if the room exists
	h.mu.Lock()
	room, exists := h.Rooms[roomID]
	h.mu.Unlock()
	if !exists {
		log.Printf("Error: Room not found")
		msg.Conn.WriteJSON(map[string]interface{}{
			"type":   "error",
			"roomID": roomID,
			"msg":    "Room not found",
		})
		return
	}

	// Deal a card
	card, ok := dealCard(room, playerID)
	if !ok {
		log.Printf("Error: Not enough cards in deck for room %s", roomID)
		return
	}

	// Get players in the room
	playersRes := db.RedisClient.HGet(db.Ctx, "room:"+roomID, "players")
	if playersRes.Err() != nil || playersRes.Val() == "" {
		log.Printf("Error: No players found for room %s", roomID)
		return
	}
	players := strings.Split(playersRes.Val(), ",")

	// Set the turn for the next player
	for _, player := range players {
		if player != playerID {
			err := db.RedisClient.HSet(db.Ctx, "room:"+roomID, "turn", player).Err()
			if err != nil {
				log.Printf("Error setting turn: %v", err)
				return
			}
			break
		}
	}

	// Broadcast the hit card
	h.broadcastRoom(roomID, map[string]interface{}{
		"type":      "hit",
		"forPlayer": playerID,
		"card":      card,
	})
}
