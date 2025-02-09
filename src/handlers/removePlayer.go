package handlers

import (
	"blackjack/src/db"
	"log"
	"strings"
)

func removePlayer(players string, playerID string, roomID string) {
	// Remove the player from the players list in Redis (comma-separated string)
	playersList := strings.Split(players, ",")
	var updatedPlayersList []string
	for _, p := range playersList {
		if p != playerID {
			updatedPlayersList = append(updatedPlayersList, p)
		}
	}
	updatedPlayers := strings.Join(updatedPlayersList, ",")
	err := db.RedisClient.HSet(db.Ctx, "room:"+roomID, "players", updatedPlayers).Err()
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

	// Optionally, if the room is now empty, you can delete the room data from Redis
	if len(updatedPlayers) == 0 {
		err = db.RedisClient.Del(db.Ctx, "room:"+roomID).Err()
		if err != nil {
			log.Println("Error deleting room from Redis:", err)
		}
	}
}
