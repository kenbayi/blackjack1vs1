package model

type CreateRoomParams struct {
	UserID string
	Bet    int
}

type JoinRoomParams struct {
	UserID string
	RoomID string
	Bet    int
}

type LeaveRoomParams struct {
	UserID string
	RoomID string
}

type PlayerLeftRoomNotification struct {
	Players []string
	RoomID  string
	Message string
}
type UpdateRoomNotification struct {
	Action  string
	RoomID  string
	Players []string
}
type RemoveRoomNotification struct {
	RoomID  string
	Action  string
	Players []string
	Status  string
	Bet     int
}

type PlayerReadyParams struct {
	UserID  string
	RoomID  string
	IsReady bool
}

type HitParams struct {
	UserID string
	RoomID string
}

type StandParams struct {
	UserID string
	RoomID string
}
