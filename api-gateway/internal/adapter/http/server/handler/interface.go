package handler

import (
	"api-gateway/internal/model"
	"context"
)

type UserUsecase interface {
	Register(ctx context.Context, user model.User) (int64, error)
	Login(ctx context.Context, user model.User) (model.Token, error)
	RefreshToken(ctx context.Context, token model.Token) (model.Token, error)
	DeleteByID(ctx context.Context) (model.UserUpdateData, error)
	UpdateUsername(ctx context.Context, user model.UserUpdateData) (model.UserUpdateData, error)
	UpdateEmailRequest(ctx context.Context, user model.UserUpdateData) error
	ConfirmEmailChange(ctx context.Context, req model.RequestToChange) error
	ChangePassword(ctx context.Context, user model.User) error
	RequestPasswordReset(ctx context.Context, user model.User) error
	ResetPassword(ctx context.Context, req model.RequestToChange) error
}
