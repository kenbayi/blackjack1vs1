package server

import "auth_svc/internal/adapter/grpc/server/frontend"

type CustomerUsecase interface {
	frontend.UserUsecase
}
