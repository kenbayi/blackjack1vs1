package handlers

import (
	"blackjack/src/db"
	"blackjack/src/models"
	"context"
	"log"
	"strconv"
	"strings"
)

func (h *Hub) endGame(roomID string, winnerID string) {
	h.mu.Lock()
	_, exists := h.Rooms[roomID]
	h.mu.Unlock()

	if !exists {
		log.Printf("Error: Room not found for endGame")
		return
	}

	ctx := context.Background()

	// Retrieve final hands
	playersRes := db.RedisClient.HGet(db.Ctx, "room:"+roomID, "players")
	players := strings.Split(playersRes.Val(), ",")
	if len(players) != 2 {
		return
	}
	p1, p2 := players[0], players[1]

	hand1 := db.RedisClient.LRange(db.Ctx, "room:"+roomID+":hand:"+p1, 0, -1).Val()
	hand2 := db.RedisClient.LRange(db.Ctx, "room:"+roomID+":hand:"+p2, 0, -1).Val()

	score1 := calculateScore(hand1)
	score2 := calculateScore(hand2)

	// Get the bet amount from Redis
	bet := db.RedisClient.HGet(db.Ctx, "room:"+roomID, "bet").Val()
	if bet == "" {
		log.Println("Error: Bet amount not found in Redis")
		return
	}

	// Determine loser
	loserID := p2
	if winnerID == p2 {
		loserID = p1
	}

	err := models.UpdatePlayerBalances(db.PostgresDB, ctx, bet, winnerID, loserID)
	if err != nil {
		log.Println("Failed to update balances:", err)
		return
	}

	err = models.InsertGameRoom(db.PostgresDB, ctx, roomID, p1, p2, winnerID)
	if err != nil {
		log.Println("Failed to insert game room record:", err)
	}

	// Reset game state in Redis
	db.RedisClient.HSet(db.Ctx, "room:"+roomID, "status", "waiting")
	db.RedisClient.HSet(db.Ctx, "room:"+roomID, "ready."+p1, "true")
	db.RedisClient.HSet(db.Ctx, "room:"+roomID, "ready."+p2, "true")
	db.RedisClient.HSet(db.Ctx, "room:"+roomID, "turn", "")
	db.RedisClient.Del(db.Ctx, "room:"+roomID+":hand:"+p1)
	db.RedisClient.Del(db.Ctx, "room:"+roomID+":hand:"+p2)
	db.RedisClient.HSet(db.Ctx, "room:"+roomID, "score."+p1, "0")
	db.RedisClient.HSet(db.Ctx, "room:"+roomID, "score."+p2, "0")

	// Validate if the player has enough money
	playerMoney, err := models.GetPlayerBalance(db.PostgresDB, p1)
	betInt, err1 := strconv.Atoi(bet)
	if err1 != nil {
		log.Printf("Error converting bet to int")
		return
	}
	if err != nil || playerMoney < betInt {
		removePlayer(playersRes.Val(), p1, roomID)
		return
	}
	playerMoney2, err := models.GetPlayerBalance(db.PostgresDB, p2)
	betInt, err2 := strconv.Atoi(bet)
	if err2 != nil {
		log.Printf("Error converting bet to int")
		return
	}
	if err != nil || playerMoney2 < betInt {
		removePlayer(playersRes.Val(), p2, roomID)
		return
	}

	// Broadcast game result
	h.broadcastRoom(roomID, map[string]interface{}{
		"type":   "game_end",
		"roomID": roomID,
		"winner": winnerID,
		"scores": map[string]int{p1: score1, p2: score2},
		"hands":  map[string][]string{p1: hand1, p2: hand2},
	})

	// Notify players to press "Ready" for the next round
	h.broadcastRoom(roomID, map[string]interface{}{
		"type": "game_waiting",
		"msg":  "Both players need to press 'Ready' to start the next round.",
	})
}
