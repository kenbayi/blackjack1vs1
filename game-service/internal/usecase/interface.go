package usecase

import (
	"context"
	"game_svc/internal/model"
)

// RoomStateRepository отвечает за управление состоянием комнаты в Redis.
type RoomStateRepository interface {
	// GetAllRoomFields получает все поля из хеша комнаты.
	GetAllRoomFields(ctx context.Context, roomID string) (map[string]string, error)

	// UpdatePlayerList обновляет поле "players" в хеше комнаты.
	UpdatePlayerList(ctx context.Context, roomID string, updatedPlayersStr string) error

	// DeletePlayerSpecificFields удаляет набор полей, относящихся к конкретному игроку, из хеша комнаты.
	// Например: "readyStatus.playerID", "scores.playerID", "hands.playerID".
	DeletePlayerSpecificFields(ctx context.Context, roomID string, playerID string, fields []string) error

	// ResetPlayerState сбрасывает состояние игрока (score, readyStatus) в хеше комнаты.
	ResetPlayerState(ctx context.Context, roomID string, playerID string) error

	// DeleteRoom полностью удаляет хеш комнаты из Redis.
	DeleteRoom(ctx context.Context, roomID string) error

	// SetRoomField устанавливает значение одного поля в хеше комнаты (может понадобиться для статуса).
	SetRoomField(ctx context.Context, roomID string, field string, value interface{}) error

	AddJoiningPlayer(ctx context.Context, roomID string, joiningUserID string, updatedPlayersStr string) error

	SaveRoom(ctx context.Context, room *model.Room) error
}

type GameEventStorage interface {
	PushGameEnd(ctx context.Context, results *model.Result, bet int64) error
}

type ClientPresenter interface {
	AddBalance(ctx context.Context, request model.User) (model.User, error)
	SubtractBalance(ctx context.Context, request model.User) (model.User, error)
	Get(ctx context.Context, id int64) (*model.User, error)
	GetRating(ctx context.Context, id int64) (*model.User, error)
}

type MatchmakingPoolRepo interface {
	AddToPool(ctx context.Context, userID string, mmr int64) error
	FindOpponent(ctx context.Context, userID string, mmr int64, mmrRange int64) (*model.Opponent, error)
	RemoveFromPool(ctx context.Context, userIDs ...string) error
}

type RoomUseCase interface {
	JoinRoom(params model.JoinRoomParams) (*model.Room, error)
	CreateRoom(params model.CreateRoomParams) (*model.Room, error)
}
