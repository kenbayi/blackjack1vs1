package dto

import (
	svc "game_svc/internal/adapter/grpc/server/frontend/proto/user"
	"game_svc/internal/model"
)

func FromGRPCClientGetResponse(resp *svc.GetBalanceResponse) *model.User {
	return &model.User{
		Balance: &resp.Balance,
	}
}
