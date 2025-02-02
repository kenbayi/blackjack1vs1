package handlers

import (
	"blackjack/src/db"
	"log"
	"strings"
)

func (h *Hub) startGame(roomID string) {
	h.mu.Lock()
	room, exists := h.Rooms[roomID]
	h.mu.Unlock()

	if !exists {
		return
	}
	playersRes := db.RedisClient.HGet(db.Ctx, "room:"+roomID, "players")
	players := strings.Split(playersRes.Val(), ",")
	if len(players) != 2 {
		log.Printf("Error: room %s does not have exactly 2 players", roomID)
		return
	}

	err := db.RedisClient.HSet(db.Ctx, "room:"+roomID, "turn", players[0]).Err()
	if err != nil {
		log.Println("Error setting turn for player:", err)
		return
	}

	p1, p2 := players[0], players[1]

	// Deal two cards to each player
	card1, ok1 := dealCard(room, p1)
	card2, ok2 := dealCard(room, p1)
	card3, ok3 := dealCard(room, p2)
	card4, ok4 := dealCard(room, p2)

	if !ok1 || !ok2 || !ok3 || !ok4 {
		log.Printf("Error: Not enough cards in deck for room %s", roomID)
		return
	}

	// Broadcast game start message with initial hands
	h.broadcastRoom(roomID, map[string]interface{}{
		"type": "game_start",
		"hands": map[string][]string{
			p1: {card1, card2},
			p2: {card3, card4},
		},
		"turn": players[0],
	})
}
