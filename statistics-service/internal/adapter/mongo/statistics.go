package mongo

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"statistics/internal/adapter/mongo/dao"
	"statistics/internal/model"
	"time"
)

const (
	generalStatsCollection = "general_stats"   // Stores a single document for general game stats
	userStatsCollection    = "user_game_stats" // Stores stats per user
)

type StatisticsRepoImpl struct {
	db *mongo.Database
}

func NewStatisticsRepository(db *mongo.Database) *StatisticsRepoImpl {
	return &StatisticsRepoImpl{db: db}
}

func getGeneralStatsDocID() string {
	return "global_blackjack_stats"
}

func (r *StatisticsRepoImpl) IncrementTotalUsers(ctx context.Context) error {
	coll := r.db.Collection(generalStatsCollection)
	filter := bson.M{"_id": getGeneralStatsDocID()}
	update := bson.M{
		"$inc": bson.M{"total_users": 1},
		"$set": bson.M{"last_updated_at": time.Now()},
	}
	opts := options.Update().SetUpsert(true)

	_, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("mongo: failed to increment total users: %w", err)
	}
	return nil
}

func (r *StatisticsRepoImpl) DecrementTotalUsers(ctx context.Context) error {
	coll := r.db.Collection(generalStatsCollection)
	filter := bson.M{"_id": getGeneralStatsDocID()}
	update := bson.M{
		"$inc": bson.M{"total_users": -1},
		"$set": bson.M{"last_updated_at": time.Now()},
	}
	opts := options.Update().SetUpsert(true)

	_, err := coll.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("mongo: failed to decrement total users: %w", err)
	}
	return nil
}

func (r *StatisticsRepoImpl) UpdateStatsForGameResult(ctx context.Context, gameResult model.GameResultEventData) error {
	// 1. Update General Stats
	generalColl := r.db.Collection(generalStatsCollection)
	generalFilter := bson.M{"_id": getGeneralStatsDocID()}
	generalUpdate := bson.M{
		"$inc": bson.M{
			"total_games_played": 1,
			"total_bet_amount":   gameResult.Bet,
		},
		"$set": bson.M{"last_updated_at": time.Now()},
	}
	generalOpts := options.Update().SetUpsert(true)
	if _, err := generalColl.UpdateOne(ctx, generalFilter, generalUpdate, generalOpts); err != nil {
		return fmt.Errorf("mongo: failed to update general game stats: %w", err)
	}

	// 2. Update Stats for Player1 and Player2
	playersInGame := []model.PlayerGameResultData{}
	if gameResult.Player1.PlayerID != 0 {
		playersInGame = append(playersInGame, gameResult.Player1)
	}
	if gameResult.Player2.PlayerID != 0 && gameResult.Player2.PlayerID != gameResult.Player1.PlayerID { // Ensure distinct players
		playersInGame = append(playersInGame, gameResult.Player2)
	}

	for _, playerData := range playersInGame {
		userColl := r.db.Collection(userStatsCollection)
		userFilter := bson.M{"user_id": playerData.PlayerID}

		// Fetch current user stats to update streaks
		var currentUserStatsDAO dao.UserGameStatsDAO
		err := userColl.FindOne(ctx, userFilter).Decode(&currentUserStatsDAO)

		// Initialize if not found (new player for stats)
		if errors.Is(err, mongo.ErrNoDocuments) {
			currentUserStatsDAO = dao.UserGameStatsDAO{
				UserID:        playerData.PlayerID,
				WinStreak:     0,
				LossStreak:    0,
				GamesPlayed:   0,
				GamesWon:      0,
				GamesLost:     0,
				GamesDrawn:    0,
				TotalBet:      0,
				TotalWinnings: 0,
				TotalLosses:   0,
			}
		} else if err != nil {
			log.Printf("mongo: Failed to fetch current stats for player %d to update streak: %v", playerData.PlayerID, err)
			return fmt.Errorf("mongo: failed to fetch current stats for player %d: %w", playerData.PlayerID, err)
		}

		var gamesWonInc, gamesLostInc, gamesDrawnInc int64
		var winningsInc, lossesInc int64
		var playerBetAmount int64 = gameResult.Bet

		currentWinStreak := currentUserStatsDAO.WinStreak
		currentLossStreak := currentUserStatsDAO.LossStreak

		isWinner := playerData.PlayerID == gameResult.WinnerID
		isLoser := playerData.PlayerID == gameResult.LoserID
		isDraw := gameResult.WinnerID == 0

		if isDraw {
			gamesDrawnInc = 1
			currentWinStreak = 0  // Reset win streak on a draw
			currentLossStreak = 0 // Reset loss streak on a draw
		} else if isWinner {
			gamesWonInc = 1
			winningsInc = gameResult.Bet
			currentWinStreak++    // Increment win streak
			currentLossStreak = 0 // Reset loss streak
		} else if isLoser { // isLoser or just `else` if not winner and not draw
			gamesLostInc = 1
			lossesInc = gameResult.Bet
			currentLossStreak++  // Increment loss streak
			currentWinStreak = 0 // Reset win streak
		}

		userUpdate := bson.M{
			"$inc": bson.M{
				"games_played":   1,
				"games_won":      gamesWonInc,
				"games_lost":     gamesLostInc,
				"games_drawn":    gamesDrawnInc,
				"total_bet":      playerBetAmount,
				"total_winnings": winningsInc,
				"total_losses":   lossesInc,
			},
			"$set": bson.M{
				"last_game_played_at": gameResult.CreatedAt,
				"win_streak":          currentWinStreak,  // Set the new win streak
				"loss_streak":         currentLossStreak, // Set the new loss streak
			},
			"$setOnInsert": bson.M{
				"user_id": playerData.PlayerID,
			},
		}

		userOpts := options.Update().SetUpsert(true)
		if _, err := userColl.UpdateOne(ctx, userFilter, userUpdate, userOpts); err != nil {
			log.Printf("mongo: Failed to update user game stats for player %d: %v", playerData.PlayerID, err)
			return fmt.Errorf("mongo: failed to update stats for player %d: %w", playerData.PlayerID, err)
		}
	}
	return nil
}

func (r *StatisticsRepoImpl) RepoGetGeneralGameStats(ctx context.Context) (model.GeneralGameStats, error) {
	coll := r.db.Collection(generalStatsCollection)
	filter := bson.M{"_id": getGeneralStatsDocID()}
	var statsDAO dao.GeneralStatsDAO

	err := coll.FindOne(ctx, filter).Decode(&statsDAO)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.GeneralGameStats{LastUpdatedAt: time.Now()}, nil
		}
		return model.GeneralGameStats{}, fmt.Errorf("mongo: failed to get general stats: %w", err)
	}
	return dao.ToGeneralStatsModel(statsDAO), nil
}

func (r *StatisticsRepoImpl) RepoGetUserGameStats(ctx context.Context, userID int64) (model.UserGameStats, error) {
	coll := r.db.Collection(userStatsCollection)
	filter := bson.M{"user_id": userID}
	var statsDAO dao.UserGameStatsDAO

	err := coll.FindOne(ctx, filter).Decode(&statsDAO)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Return UserGameStats with UserID and zero values for calculable fields
			m := model.UserGameStats{UserID: userID}
			if m.GamesPlayed > 0 { // Should be 0 here
				// This calculation will result in 0 anyway
			} else {
				m.WinRate = 0.0
				m.LossRate = 0.0
			}
			return m, nil
		}
		return model.UserGameStats{}, fmt.Errorf("mongo: failed to get stats for user %d: %w", userID, err)
	}
	return dao.ToUserGameStatsModel(statsDAO), nil
}

func (r *StatisticsRepoImpl) RepoGetLeaderboard(ctx context.Context, leaderboardType string, limit int) (model.Leaderboard, error) {
	coll := r.db.Collection(userStatsCollection)
	var resultsDAO []dao.UserGameStatsDAO

	var sortFieldBSON string
	switch leaderboardType {
	case "top_wins":
		sortFieldBSON = "games_won"
	case "top_earnings":
		sortFieldBSON = "total_winnings"
	case "top_win_streak":
		sortFieldBSON = "win_streak"
	default:
		return model.Leaderboard{}, fmt.Errorf("unsupported leaderboard type: %s", leaderboardType)
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: sortFieldBSON, Value: -1}})
	findOptions.SetLimit(int64(limit))

	cursor, err := coll.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return model.Leaderboard{}, fmt.Errorf("mongo: failed to query leaderboard: %w", err)
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &resultsDAO); err != nil {
		return model.Leaderboard{}, fmt.Errorf("mongo: failed to decode leaderboard results: %w", err)
	}

	leaderboard := model.Leaderboard{
		Type:    leaderboardType,
		Entries: make([]model.LeaderboardEntry, len(resultsDAO)),
	}

	for i, userStatDAO := range resultsDAO {
		userStatModel := dao.ToUserGameStatsModel(userStatDAO)
		score := userStatModel.GamesWon // Default for "top_wins"
		if leaderboardType == "top_earnings" {
			score = userStatModel.TotalWinnings - userStatModel.TotalLosses
		} else if leaderboardType == "top_win_streak" {
			score = userStatModel.WinStreak
		}
		// TODO: Fetch Username if needed, or store it denormalized in userStatsCollection
		leaderboard.Entries[i] = model.LeaderboardEntry{
			UserID: userStatModel.UserID,
			Score:  score,
			Rank:   i + 1,
		}
	}
	return leaderboard, nil
}
