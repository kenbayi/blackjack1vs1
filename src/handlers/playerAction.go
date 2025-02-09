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
		msg.Conn.WriteJSON(map[string]interface{}{
			"type":   "warning",
			"roomID": roomID,
			"msg":    "Not your turn",
		})
		return
	}

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

	// Get updated score
	hand := db.RedisClient.LRange(db.Ctx, "room:"+roomID+":hand:"+playerID, 0, -1).Val()
	score := calculateScore(hand)
	err := db.RedisClient.HSet(db.Ctx, "room:"+roomID, "scores."+playerID, score).Err()
	if err != nil {
		log.Printf("Error setting scores for player %s in room %s: %v", playerID, roomID, err)
		return
	}

	// Get players
	playersRes := db.RedisClient.HGet(db.Ctx, "room:"+roomID, "players")
	players := strings.Split(playersRes.Val(), ",")
	if len(players) != 2 {
		log.Printf("Error: room %s does not have exactly 2 players", roomID)
		return
	}
	p1, p2 := players[0], players[1]

	// Determine opponent
	opponentID := p2
	if playerID == p2 {
		opponentID = p1
	}

	// Broadcast the hit event
	h.broadcastRoom(roomID, map[string]interface{}{
		"type":      "hit",
		"forPlayer": playerID,
		"card":      card,
		"score":     score,
	})

	// If player busts, they immediately lose & the opponent wins
	if score > 21 {
		h.broadcastRoom(roomID, map[string]interface{}{
			"type":      "busted",
			"forPlayer": playerID,
			"msg":       "Player busted!",
		})

		// Call game end, declaring opponent as the winner
		h.endGame(roomID, opponentID)
		return
	}

	// If player didn't bust, switch turns
	db.RedisClient.HSet(db.Ctx, "room:"+roomID, "turn", opponentID)

	h.broadcastRoom(roomID, map[string]interface{}{
		"type": "turn",
		"turn": opponentID,
	})
}

func (h *Hub) standCard(msg Message) {
	content := msg.Content.(map[string]interface{})
	playerID := content["playerID"].(string)
	roomID := content["roomID"].(string)

	turn := db.RedisClient.HGet(db.Ctx, "room:"+roomID, "turn").Val()
	if turn != playerID {
		log.Printf("Error: not your turn")
		msg.Conn.WriteJSON(map[string]interface{}{
			"type":   "warning",
			"roomID": roomID,
			"msg":    "Not your turn",
		})
		return
	}

	h.mu.Lock()
	_, exists := h.Rooms[roomID]
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

	// Get players
	playersRes := db.RedisClient.HGet(db.Ctx, "room:"+roomID, "players")
	players := strings.Split(playersRes.Val(), ",")
	if len(players) != 2 {
		log.Printf("Error: room %s does not have exactly 2 players", roomID)
		return
	}
	p1, p2 := players[0], players[1]

	//switch turn
	opponentID := p2
	if playerID == p2 {
		opponentID = p1
	}
	db.RedisClient.HSet(db.Ctx, "room:"+roomID, "turn", opponentID)

	// Mark player as stood
	db.RedisClient.HSet(db.Ctx, "room:"+roomID, "stood."+playerID, "1")

	// Get scores
	hand1 := db.RedisClient.LRange(db.Ctx, "room:"+roomID+":hand:"+p1, 0, -1).Val()
	hand2 := db.RedisClient.LRange(db.Ctx, "room:"+roomID+":hand:"+p2, 0, -1).Val()
	score1 := calculateScore(hand1)
	score2 := calculateScore(hand2)

	err1 := db.RedisClient.HSet(db.Ctx, "room:"+roomID, "scores."+p1, score1).Err()
	if err1 != nil {
		log.Printf("Error setting scores for player %s in room %s: %v", p1, roomID, err1)
		return
	}
	err2 := db.RedisClient.HSet(db.Ctx, "room:"+roomID, "scores."+p2, score2).Err()
	if err2 != nil {
		log.Printf("Error setting scores for player %s in room %s: %v", p1, roomID, err1)
		return
	}
	// Check if both players have stood
	p1Stood := db.RedisClient.HGet(db.Ctx, "room:"+roomID, "stood."+p1).Val() == "1"
	p2Stood := db.RedisClient.HGet(db.Ctx, "room:"+roomID, "stood."+p2).Val() == "1"

	// Broadcast stand event
	h.broadcastRoom(roomID, map[string]interface{}{
		"type":      "stand",
		"forPlayer": playerID,
		"scores": map[string]int{
			p1: score1,
			p2: score2,
		},
	})
	h.broadcastRoom(roomID, map[string]interface{}{
		"type": "turn",
		"turn": opponentID,
	})
	// If both players have stood, end the game and declare a winner
	if p1Stood && p2Stood {
		var winner string
		if score1 > score2 {
			winner = p1
		} else if score2 > score1 {
			winner = p2
		} else {
			winner = "" // Tie
		}
		h.endGame(roomID, winner)
	}
}
