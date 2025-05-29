package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"statistics/internal/model"
)

const (
	gameHistoryCollection = "game_history"
)

// GameHistoryRepoImpl implements usecase.GameHistoryRepository
type GameHistoryRepoImpl struct {
	db *mongo.Database
}

// NewGameHistoryRepository creates a new instance of GameHistoryRepoImpl.
func NewGameHistoryRepository(db *mongo.Database) *GameHistoryRepoImpl {
	return &GameHistoryRepoImpl{db: db}
}

// InsertGame saves a single game history entry to MongoDB.
func (r *GameHistoryRepoImpl) InsertGame(ctx context.Context, gameHistoryEntry model.GameHistory) error {
	coll := r.db.Collection(gameHistoryCollection)

	_, err := coll.InsertOne(ctx, gameHistoryEntry)
	if err != nil {
		log.Printf("mongo: Failed to insert game history for RoomID %s: %v", gameHistoryEntry.RoomID, err)
		return fmt.Errorf("mongo: failed to insert game history for RoomID %s: %w", gameHistoryEntry.RoomID, err)
	}

	log.Printf("mongo: Successfully inserted game history for RoomID %s, GameID: %s", gameHistoryEntry.RoomID, gameHistoryEntry.ID)
	return nil
}
