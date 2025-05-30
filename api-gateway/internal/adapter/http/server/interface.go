package server

import (
	"api-gateway/internal/adapter/http/server/handler"
)

type UserUsecase interface {
	handler.UserUsecase
}

type UserProfileUsecase interface {
	handler.UserProfileUsecase
}

type StatisticsUsecase interface {
	handler.StatisticsUsecase
}
