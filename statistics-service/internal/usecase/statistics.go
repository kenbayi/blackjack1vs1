package usecase

import (
	"context"
	"fmt"
	"log"
	"statistics/internal/model"
	"time"
)

type StatisticsUseCase struct {
	repo            StatisticsRepository
	redis           StatisticsRedisCache
	gameHistoryRepo GameHistoryRepository
	cacheTTL        time.Duration
}

func NewStatisticsUseCase(
	repo StatisticsRepository,
	redisCache StatisticsRedisCache,
	ghr GameHistoryRepository,
) *StatisticsUseCase {
	return &StatisticsUseCase{
		repo:            repo,
		redis:           redisCache,
		gameHistoryRepo: ghr,
		cacheTTL:        15 * time.Minute,
	}
}

// --- Event Handling Methods ---

func (uc *StatisticsUseCase) HandleUserCreated(ctx context.Context, eventData model.UserCreatedEventData) error {
	log.Printf("StatisticsUseCase: Handling UserCreated event for UserID: %d", eventData.ID)
	if err := uc.repo.IncrementTotalUsers(ctx); err != nil {
		log.Printf("StatisticsUseCase: Failed to increment total users: %v", err)
		return fmt.Errorf("failed to increment total users: %w", err)
	}

	if err := uc.redis.DeleteGeneralGameStats(ctx); err != nil {
		log.Printf("StatisticsUseCase: Warning - Failed to delete general stats cache after user creation: %v", err)
	}

	log.Printf("StatisticsUseCase: UserCreated event processed for UserID: %d", eventData.ID)
	return nil
}

func (uc *StatisticsUseCase) HandleUserDeleted(ctx context.Context, eventData model.UserDeletedEventData) error {
	log.Printf("StatisticsUseCase: Handling UserDeleted event for UserID: %d", eventData.ID)

	if err := uc.repo.DecrementTotalUsers(ctx); err != nil {
		log.Printf("StatisticsUseCase: Failed to decrement total users: %v", err)
		return fmt.Errorf("failed to decrement total users: %w", err)
	}

	if err := uc.redis.DeleteGeneralGameStats(ctx); err != nil {
		log.Printf("StatisticsUseCase: Warning - Failed to delete general stats cache after user deletion: %v", err)
	}

	if err := uc.redis.DeleteUserGameStats(ctx, eventData.ID); err != nil {
		log.Printf("StatisticsUseCase: Warning - Failed to delete user stats cache for UserID %d: %v", eventData.ID, err)
	}

	log.Printf("StatisticsUseCase: UserDeleted event processed for UserID: %d", eventData.ID)
	return nil
}

// HandleGameResult now accepts model.GameResultEventData
func (uc *StatisticsUseCase) HandleGameResult(ctx context.Context, eventData model.GameResultEventData) error {
	log.Printf("StatisticsUseCase: Handling GameResult event for RoomID: %s, Winner: %d", eventData.RoomID, eventData.WinnerID)

	// 1. Update aggregated statistics (general and per-player)
	if err := uc.repo.UpdateStatsForGameResult(ctx, eventData); err != nil {
		log.Printf("StatisticsUseCase: Failed to update stats for game result (RoomID: %s): %v", eventData.RoomID, err)
		return fmt.Errorf("failed to update stats for game result: %w", err)
	}

	// 2. Record game history
	gameHistoryEntry := model.GameHistory{
		RoomID:       eventData.RoomID,
		WinnerID:     eventData.WinnerID,
		LoserID:      eventData.LoserID,
		BetAmount:    eventData.Bet,
		GameEndedAt:  eventData.CreatedAt,
		Player1ID:    eventData.Player1.PlayerID,
		Player1Hand:  eventData.Player1.FinalHand,
		Player1Score: eventData.Player1.FinalScore,
		Player2ID:    eventData.Player2.PlayerID,
		Player2Hand:  eventData.Player2.FinalHand,
		Player2Score: eventData.Player2.FinalScore,
	}

	if err := uc.gameHistoryRepo.InsertGame(ctx, gameHistoryEntry); err != nil {
		log.Printf("StatisticsUseCase: Failed to insert game history for RoomID %s: %v", eventData.RoomID, err)
	}

	// 3. Invalidate/clear relevant caches
	if err := uc.redis.DeleteGeneralGameStats(ctx); err != nil {
		log.Printf("StatisticsUseCase: Warning - Failed to delete general stats cache: %v", err)
	}

	// Player IDs are int64 in eventData.Player1 and eventData.Player2
	if err := uc.redis.DeleteUserGameStats(ctx, eventData.Player1.PlayerID); err != nil {
		log.Printf("StatisticsUseCase: Warning - Failed to delete user stats cache for PlayerID %d: %v", eventData.Player1.PlayerID, err)
	}

	if err := uc.redis.DeleteUserGameStats(ctx, eventData.Player2.PlayerID); err != nil {
		log.Printf("StatisticsUseCase: Warning - Failed to delete user stats cache for PlayerID %d: %v", eventData.Player2.PlayerID, err)
	}

	err := uc.redis.DeleteLeaderboard(ctx, "top_wins")
	if err != nil {
		return err
	}

	log.Printf("StatisticsUseCase: GameResult event processed for RoomID: %s", eventData.RoomID)
	return nil
}

func (uc *StatisticsUseCase) GetGeneralGameStats(ctx context.Context) (model.GeneralGameStats, error) {
	log.Printf("StatisticsUseCase: GetGeneralGameStats called")
	cachedStats, err := uc.redis.RepoGetGeneralGameStats(ctx)
	if err == nil {
		log.Printf("StatisticsUseCase: GetGeneralGameStats cache hit")
		return cachedStats, nil
	}
	log.Printf("StatisticsUseCase: GetGeneralGameStats cache miss/error: %v", err)
	repoStats, err := uc.repo.RepoGetGeneralGameStats(ctx)

	if err != nil {
		log.Printf("StatisticsUseCase: Error from repository GetGeneralGameStats: %v", err)
		return model.GeneralGameStats{}, fmt.Errorf("repository error: %w", err)
	}

	go func(statsToCache model.GeneralGameStats) {
		bgCtx := context.Background()
		if errCache := uc.redis.SetGeneralGameStats(bgCtx, statsToCache, uc.cacheTTL); errCache != nil {
			log.Printf("StatisticsUseCase: Error setting general stats to cache: %v", errCache)
		} else {
			log.Printf("StatisticsUseCase: General stats set to cache")
		}
	}(repoStats)
	return repoStats, nil
}

func (uc *StatisticsUseCase) GetUserGameStats(ctx context.Context, userID int64) (model.UserGameStats, error) {
	log.Printf("StatisticsUseCase: GetUserGameStats called for UserID: %d", userID)
	cachedStats, err := uc.redis.RepoGetUserGameStats(ctx, userID)

	if err == nil {
		log.Printf("StatisticsUseCase: GetUserGameStats cache hit for UserID: %d", userID)
		return cachedStats, nil
	}
	log.Printf("StatisticsUseCase: GetUserGameStats cache miss/error for UserID %d: %v", userID, err)
	repoStats, err := uc.repo.RepoGetUserGameStats(ctx, userID)

	if err != nil {
		log.Printf("StatisticsUseCase: Error from repository GetUserGameStats for UserID %d: %v", userID, err)
		return model.UserGameStats{}, fmt.Errorf("repository error for UserID %d: %w", userID, err)
	}

	// Calculate WinRate and LossRate
	if repoStats.GamesPlayed > 0 {
		repoStats.WinRate = float64(repoStats.GamesWon) / float64(repoStats.GamesPlayed)
		repoStats.LossRate = float64(repoStats.GamesLost) / float64(repoStats.GamesPlayed)
	} else {
		repoStats.WinRate = 0.0
		repoStats.LossRate = 0.0
	}

	go func(statsToCache model.UserGameStats) {
		bgCtx := context.Background()
		if errCache := uc.redis.SetUserGameStats(bgCtx, userID, statsToCache, uc.cacheTTL); errCache != nil {
			log.Printf("StatisticsUseCase: Error setting user stats to cache for UserID %d: %v", userID, errCache)
		} else {
			log.Printf("StatisticsUseCase: User stats for UserID %d set to cache", userID)
		}
	}(repoStats)
	return repoStats, nil

}

func (uc *StatisticsUseCase) GetLeaderboard(ctx context.Context, leaderboardType string, limit int) (model.Leaderboard, error) {
	log.Printf("StatisticsUseCase: GetLeaderboard called for Type: %s, Limit: %d", leaderboardType, limit)
	cachedLeaderboard, err := uc.redis.RepoGetLeaderboard(ctx, leaderboardType, limit)

	if err == nil {
		log.Printf("StatisticsUseCase: GetLeaderboard cache hit for Type: %s", leaderboardType)
		return cachedLeaderboard, nil
	}
	log.Printf("StatisticsUseCase: GetLeaderboard cache miss/error for Type %s: %v", leaderboardType, err)
	repoLeaderboard, err := uc.repo.RepoGetLeaderboard(ctx, leaderboardType, limit)

	if err != nil {
		log.Printf("StatisticsUseCase: Error from repository GetLeaderboard for Type %s: %v", leaderboardType, err)
		return model.Leaderboard{}, fmt.Errorf("repository error for leaderboard Type %s: %w", leaderboardType, err)
	}

	go func(boardToCache model.Leaderboard) {
		bgCtx := context.Background()
		if errCache := uc.redis.SetLeaderboard(bgCtx, leaderboardType, boardToCache, uc.cacheTTL); errCache != nil {
			log.Printf("StatisticsUseCase: Error setting leaderboard to cache for Type %s: %v", leaderboardType, errCache)
		} else {
			log.Printf("StatisticsUseCase: Leaderboard for Type %s set to cache", leaderboardType)
		}
	}(repoLeaderboard)
	return repoLeaderboard, nil
}
