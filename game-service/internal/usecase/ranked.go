package usecase

import (
	"context"
	"fmt"
	"game_svc/internal/model"
	"log"
	"strconv"
)

type RankedUseCase struct {
	poolRepo        MatchmakingPoolRepo
	clientPresenter ClientPresenter
	roomUsecase     RoomUseCase
}

func NewRankedUseCase(poolRepo MatchmakingPoolRepo, presenter ClientPresenter, roomUsecase RoomUseCase) *RankedUseCase {
	return &RankedUseCase{
		poolRepo:        poolRepo,
		clientPresenter: presenter,
		roomUsecase:     roomUsecase,
	}
}

func (uc *RankedUseCase) FindMatch(userID string) (*model.Match, error) {
	ctx := context.Background()
	// 1. Fetch user's rating (MMR) from the Statistics Service
	userIDint, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, err
	}
	userData, err := uc.clientPresenter.GetRating(ctx, userIDint)
	if err != nil {
		return nil, fmt.Errorf("could not fetch user rating: %w", err)
	}

	// 2. Try to find an existing opponent in the pool
	opponent, err := uc.poolRepo.FindOpponent(ctx, userID, *userData.Rating, 100) // Example: 100 MMR range
	if err != nil {
		return nil, fmt.Errorf("error while searching for opponent: %w", err)
	}

	// 3. If NO opponent is found, add the current user to the pool and wait.
	if opponent == nil {
		err := uc.poolRepo.AddToPool(ctx, userID, *userData.Rating)
		if err != nil {
			return nil, fmt.Errorf("failed to add user to matchmaking pool: %w", err)
		}
		log.Printf("User %s added to matchmaking pool with MMR %d. Waiting for opponent.", userID, userData)
		return nil, nil
	}

	log.Printf("Match found for user %s (MMR %d) with opponent %s (MMR %d)!", userID, userData, opponent.ID, opponent.MMR)

	// a. Remove both players from the pool
	if err := uc.poolRepo.RemoveFromPool(ctx, userID, opponent.ID); err != nil {
		log.Printf("CRITICAL: Failed to remove players from pool after match was found: %v", err)
		return nil, err
	}

	createParams := model.CreateRoomParams{
		Bet:    2500,
		UserID: userID,
	}
	createdRoom, err := uc.roomUsecase.CreateRoom(createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create room for ranked match: %w", err)
	}

	// c. The second player automatically joins.
	joinParams := model.JoinRoomParams{
		RoomID: createdRoom.ID,
		UserID: opponent.ID,
		Bet:    2500,
	}
	finalRoom, err := uc.roomUsecase.JoinRoom(joinParams)
	if err != nil {
		return nil, fmt.Errorf("opponent failed to join ranked match room: %w", err)
	}

	log.Printf("Successfully created ranked room %s for players %s and %s", finalRoom.ID, userID, opponent.ID)

	// 5. Return the match details so the handler can notify both users.
	match := &model.Match{
		RoomID:  finalRoom.ID,
		Players: []string{userID, opponent.ID},
	}
	return match, nil
}
