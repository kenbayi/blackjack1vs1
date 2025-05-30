package redis

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"game_svc/internal/model"
	"game_svc/pkg/redis"
	go_redis "github.com/redis/go-redis/v9"
)

const matchmakingPoolKey = "matchmaking:pool"

type RankedRepoImpl struct {
	client *redis.Client
}

func NewRankedRepoImpl(client *redis.Client) *RankedRepoImpl {
	return &RankedRepoImpl{client: client}
}

// AddToPool adds a user to the matchmaking pool sorted set.
func (r *RankedRepoImpl) AddToPool(ctx context.Context, userID string, mmr int64) error {
	err := r.client.Unwrap().ZAdd(ctx, matchmakingPoolKey, go_redis.Z{
		Score:  float64(mmr),
		Member: userID,
	}).Err()

	if err != nil {
		return fmt.Errorf("redis ZADD failed for user %s in matchmaking pool: %w", userID, err)
	}

	log.Printf("Redis: User %s with MMR %d added to matchmaking pool.", userID, mmr)
	return nil
}

// RemoveFromPool removes one or more users from the matchmaking pool.
func (r *RankedRepoImpl) RemoveFromPool(ctx context.Context, userIDs ...string) error {
	if len(userIDs) == 0 {
		return nil
	}
	members := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		members[i] = id
	}

	err := r.client.Unwrap().ZRem(ctx, matchmakingPoolKey, members...).Err()
	if err != nil {
		return fmt.Errorf("redis ZREM failed for matchmaking pool: %w", err)
	}

	log.Printf("Redis: Removed users %v from matchmaking pool.", userIDs)
	return nil
}

// FindOpponent transactionally finds and removes a suitable opponent from the pool.
func (r *RankedRepoImpl) FindOpponent(ctx context.Context, searchingUserID string, searchingUserMMR int64, mmrRange int64) (*model.Opponent, error) {
	minMMR := strconv.Itoa(int(searchingUserMMR - mmrRange))
	maxMMR := strconv.Itoa(int(searchingUserMMR + mmrRange))

	// LOGIC:
	// 1. Find all users in the given MMR range.
	// 2. Iterate through them.
	// 3. If an opponent is found that is NOT the user themselves:
	//    a. Remove that opponent from the pool.
	//    b. Return the opponent's ID and MMR.
	// 4. If no suitable opponent is found, return nil.
	script := `
		local pool_key = KEYS[1]
		local min_mmr = ARGV[1]
		local max_mmr = ARGV[2]
		local searching_user_id = ARGV[3]

		-- Find players in the specified MMR range
		local opponents = redis.call('ZRANGEBYSCORE', pool_key, min_mmr, max_mmr, 'WITHSCORES')

		if #opponents == 0 then
			return nil
		end

		-- Iterate to find the first opponent who is not the searcher
		for i=1, #opponents, 2 do
			local opponent_id = opponents[i]
			local opponent_mmr = opponents[i+1]
			
			if opponent_id ~= searching_user_id then
				-- Found a valid opponent, remove them from the pool
				redis.call('ZREM', pool_key, opponent_id)
				-- Return their ID and MMR
				return {opponent_id, opponent_mmr}
			end
		end

		-- No one else in the range
		return nil
	`

	// Execute the script
	result, err := r.client.Unwrap().Eval(ctx, script, []string{matchmakingPoolKey}, minMMR, maxMMR, searchingUserID).Result()
	if err == go_redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis Lua script for FindOpponent failed: %w", err)
	}

	resSlice, ok := result.([]interface{})
	if !ok || len(resSlice) != 2 {
		return nil, fmt.Errorf("unexpected result format from FindOpponent Lua script: %v", result)
	}

	opponentID, okID := resSlice[0].(string)
	opponentMMRStr, okMMR := resSlice[1].(string)
	if !okID || !okMMR {
		return nil, fmt.Errorf("type assertion failed for opponent data from Lua script")
	}

	opponentMMR, err := strconv.ParseInt(opponentMMRStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse opponent MMR from string '%s': %w", opponentMMRStr, err)
	}

	return &model.Opponent{
		ID:  opponentID,
		MMR: opponentMMR,
	}, nil
}
