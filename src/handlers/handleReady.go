package handlers

import (
	"blackjack/src/db"
	"log"
	"strings"
)

func (h *Hub) handleReady(msg Message) {
	content := msg.Content.(map[string]interface{})
	roomID := content["roomID"].(string)
	playerID := content["playerID"].(string)

	// Set player as ready
	res := db.RedisClient.HSet(db.Ctx, "room:"+roomID, "readyStatus."+playerID, true).Err()
	if res != nil {
		log.Printf("Error setting readyStatus for player %s in room %s: %v", playerID, roomID, res)
		return
	}

	// Check if both players are ready
	if allReady(roomID) {
		// Both players are ready, update room status to 'in_progress'
		err := db.RedisClient.HSet(db.Ctx, "room:"+roomID, "status", "in_progress").Err()
		if err != nil {
			log.Printf("Error setting room status to 'in_progress' for room %s: %v", roomID, err)
			return
		}

		// Broadcast game start message
		h.broadcastRoom(roomID, map[string]interface{}{
			"type": "game_start",
		})

		// Remove room from list of players as soon as game started
		h.broadcastAll(map[string]interface{}{
			"type":   "update_list",
			"action": "remove",
			"roomID": roomID,
		})
	}
}

func allReady(roomID string) bool {
	res := db.RedisClient.HGet(db.Ctx, "room:"+roomID, "players")
	players := strings.Split(res.Val(), ",")
	if len(players) != 2 {
		return false
	}
	p1 := db.RedisClient.HGet(db.Ctx, "room:"+roomID, "readyStatus."+players[0]).Val()
	p2 := db.RedisClient.HGet(db.Ctx, "room:"+roomID, "readyStatus."+players[1]).Val()

	if p1 == "1" && p2 == "1" {
		return true
	}

	return false
}
