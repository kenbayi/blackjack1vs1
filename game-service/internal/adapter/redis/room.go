package redis

import (
	"context"
	"fmt"
	"game_svc/internal/model"
	"log"
	"strconv"
	"strings"

	"game_svc/pkg/redis"
)

type RoomStateRepoImpl struct {
	client *redis.Client
}

func NewRoomStateRepoImpl(client *redis.Client) *RoomStateRepoImpl {
	return &RoomStateRepoImpl{client: client}
}

func roomKey(roomID string) string {
	return fmt.Sprintf("room:%s", roomID)
}

func (r *RoomStateRepoImpl) GetAllRoomFields(ctx context.Context, roomID string) (map[string]string, error) {
	key := roomKey(roomID)
	fields, err := r.client.Unwrap().HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis HGetAll for room %s failed: %w", roomID, err)
	}
	return fields, nil
}

func (r *RoomStateRepoImpl) UpdatePlayerList(ctx context.Context, roomID string, updatedPlayersStr string) error {
	key := roomKey(roomID)
	pipe := r.client.Unwrap().Pipeline()
	pipe.HSet(ctx, key, "players", updatedPlayersStr)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis HSet players for room %s failed: %w", roomID, err)
	}
	return nil
}

func (r *RoomStateRepoImpl) DeletePlayerSpecificFields(ctx context.Context, roomID string, playerID string, baseFields []string) error {
	key := roomKey(roomID)
	if len(baseFields) == 0 {
		return nil
	}

	fieldsToDelete := make([]string, len(baseFields))
	for i, baseField := range baseFields {
		fieldsToDelete[i] = fmt.Sprintf("%s.%s", baseField, playerID)
	}

	pipe := r.client.Unwrap().Pipeline()
	pipe.HDel(ctx, key, fieldsToDelete...)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis HDel player specific fields for player %s in room %s failed: %w", playerID, roomID, err)
	}
	log.Printf("Redis: Deleted fields %v for player %s in room %s", fieldsToDelete, playerID, roomID)
	return nil
}

func (r *RoomStateRepoImpl) ResetPlayerState(ctx context.Context, roomID string, playerID string) error {
	key := roomKey(roomID)
	pipe := r.client.Unwrap().Pipeline()

	// Сбрасываем поля на дефолтные значения
	pipe.HSet(ctx, key, fmt.Sprintf("scores.%s", playerID), 0)
	pipe.HSet(ctx, key, fmt.Sprintf("readyStatus.%s", playerID), 0)
	pipe.HSet(ctx, key, fmt.Sprintf("hands.%s", playerID), "nil")
	pipe.HSet(ctx, key, fmt.Sprintf("lastAction.%s", playerID), "nil")
	pipe.HSet(ctx, key, fmt.Sprintf("stood.%s", playerID), false)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis ResetPlayerState for player %s in room %s failed: %w", playerID, roomID, err)
	}
	log.Printf("Redis: Reset state for player %s in room %s", playerID, roomID)
	return nil
}

func (r *RoomStateRepoImpl) DeleteRoom(ctx context.Context, roomID string) error {
	key := roomKey(roomID)
	pipe := r.client.Unwrap().Pipeline()
	pipe.Del(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis Del room %s failed: %w", roomID, err)
	}
	log.Printf("Redis: Deleted room %s", roomID)
	return nil
}

func (r *RoomStateRepoImpl) SetRoomField(ctx context.Context, roomID string, field string, value interface{}) error {
	key := roomKey(roomID)
	pipe := r.client.Unwrap().Pipeline()
	pipe.HSet(ctx, key, field, value)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis HSet field %s for room %s failed: %w", field, roomID, err)
	}
	return nil
}

func (r *RoomStateRepoImpl) SaveRoom(ctx context.Context, room *model.Room) error {
	key := roomKey(room.ID)
	pipe := r.client.Unwrap().Pipeline()

	// Основные поля комнаты
	pipe.HSet(ctx, key, "roomID", room.ID)
	pipe.HSet(ctx, key, "status", room.Status)
	pipe.HSet(ctx, key, "bet", strconv.Itoa(room.Bet))
	pipe.HSet(ctx, key, "turn", room.CurrentTurnPlayerID)

	// Поля игроков
	if len(room.Players) > 0 {
		playerIDs := make([]string, len(room.Players))
		for i, p := range room.Players {
			playerIDs[i] = p.ID

			// Сохраняем индивидуальные состояния игроков
			readyStatus := "0"
			if p.IsReady {
				readyStatus = "1"
			}
			pipe.HSet(ctx, key, fmt.Sprintf("readyStatus.%s", p.ID), readyStatus)
			pipe.HSet(ctx, key, fmt.Sprintf("scores.%s", p.ID), strconv.Itoa(p.Score))

			var handStr string
			if len(p.Hand) == 0 {
				handStr = "nil"
			} else {
				cardStrings := make([]string, len(p.Hand))
				for j, card := range p.Hand {
					cardStrings[j] = card.Value + card.Suit
				}
				handStr = strings.Join(cardStrings, ",")
			}
			pipe.HSet(ctx, key, fmt.Sprintf("hands.%s", p.ID), handStr)
			pipe.HSet(ctx, key, fmt.Sprintf("lastAction.%s", p.ID), p.LastAction)
		}
		pipe.HSet(ctx, key, "players", strings.Join(playerIDs, ","))
	} else {
		pipe.HSet(ctx, key, "players", "") // Если игроков нет (маловероятно для CreateRoom)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis pipeline failed for SaveRoom, roomID %s: %w", room.ID, err)
	}
	log.Printf("Redis: Successfully saved room %s state", room.ID)
	return nil
}

func (r *RoomStateRepoImpl) AddJoiningPlayer(ctx context.Context, roomID string, joiningUserID string, updatedPlayersStr string) error {
	key := roomKey(roomID)
	pipe := r.client.Unwrap().Pipeline()

	// Update the players list
	pipe.HSet(ctx, key, "players", updatedPlayersStr)

	// Set default fields for the joining player
	pipe.HSet(ctx, key, fmt.Sprintf("readyStatus.%s", joiningUserID), "0") // "0" for false
	pipe.HSet(ctx, key, fmt.Sprintf("scores.%s", joiningUserID), "0")
	pipe.HSet(ctx, key, fmt.Sprintf("hands.%s", joiningUserID), "nil")      // Default empty/nil hand
	pipe.HSet(ctx, key, fmt.Sprintf("lastAction.%s", joiningUserID), "nil") // No last action yet

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis pipeline failed for AddJoiningPlayer in room %s for user %s: %w", roomID, joiningUserID, err)
	}
	log.Printf("Redis: Successfully added player %s to room %s with default fields", joiningUserID, roomID)
	return nil
}
