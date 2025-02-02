package handlers

import (
	"blackjack/src/db"
	"math/rand"
	"time"
)

// Deal a card from the deck and store it in Redis
func dealCard(room *Room, playerID string) (string, bool) {
	if len(room.Deck) == 0 {
		return "", false // No cards left in the deck
	}

	card := room.Deck[0]
	room.Deck = room.Deck[1:] // Remove dealt card from deck

	// Store card in Redis List
	key := "room:" + room.ID + ":hand:" + playerID
	db.RedisClient.RPush(db.Ctx, key, card)

	return card, true
}

// Generate a shuffled deck for the room
func generateShuffledDeck() []string {
	suits := []string{"H", "D", "C", "S"} // Hearts, Diamonds, Clubs, Spades
	values := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	deck := []string{}

	// Create a 4-deck (4 * 52 cards)
	for i := 0; i < 4; i++ {
		for _, suit := range suits {
			for _, value := range values {
				deck = append(deck, value+suit)
			}
		}
	}

	// Create a new Rand instance
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

	return deck
}
