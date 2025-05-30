package server

import (
	"api-gateway/internal/adapter/http/server/handler"
)

type UserUsecase interface {
	handler.UserUsecase
}
