package server

import (
	"game_svc/internal/adapter/ws/server/dto"
	"game_svc/internal/model"
)

type RoomUseCase interface {
	CreateRoom(params model.CreateRoomParams) (*model.Room, error)
	JoinRoom(params model.JoinRoomParams) (*model.Room, error)
	LeaveRoom(params model.LeaveRoomParams) (updatedRoom *model.Room, wasRoomDeleted bool, err error)
}

type GameUseCase interface {
	PlayerReady(params model.PlayerReadyParams) (*model.PlayerReadyResult, error)
	Hit(params model.HitParams) (*model.Result, error)
	Stand(params model.StandParams) (*model.Result, error)
	HandlePlayerDisconnect(userID string, roomID string) (*dto.DisconnectResponse, error)
}

type RankedUseCase interface {
	FindMatch(userID string) (*model.Match, error)
}
