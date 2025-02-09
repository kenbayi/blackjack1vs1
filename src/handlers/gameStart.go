package handlers

import (
	"blackjack/src/db"
	"log"
	"strconv"
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

	// Get updated scores
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
		log.Printf("Error setting scores for player %s in room %s: %v", p2, roomID, err2)
		return
	}

	// Broadcast game start message with initial hands and scores
	h.broadcastRoom(roomID, map[string]interface{}{
		"type": "game_start",
		"hands": map[string][]string{
			p1: {card1, card2},
			p2: {card3, card4},
		},
		"scores": map[string]int{
			p1: score1,
			p2: score2,
		},
		"turn": players[0],
	})
}

func calculateScore(hand []string) int {
	score := 0
	aces := 0

	for _, card := range hand {
		value := card[:len(card)-1] // Extract card value (without suit)

		switch value {
		case "A":
			aces++
			score += 11 // Initially count ace as 11
		case "K", "Q", "J":
			score += 10
		default:
			num, err := strconv.Atoi(value)
			if err == nil {
				score += num
			}
		}
	}

	// Convert Aces from 11 to 1 if needed to avoid bust
	for score > 21 && aces > 0 {
		score -= 10
		aces--
	}

	return score
}
