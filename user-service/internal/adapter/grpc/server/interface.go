package server

import "user_svc/internal/adapter/grpc/server/frontend"

type CustomerUsecase interface {
	frontend.UserUsecase
}
