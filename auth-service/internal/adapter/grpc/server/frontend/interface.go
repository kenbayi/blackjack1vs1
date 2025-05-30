package frontend

import (
	"context"

	"auth_svc/internal/model"
)

type UserUsecase interface {
	Register(ctx context.Context, request model.User) (int64, error)
	Login(ctx context.Context, email, password string) (model.Token, error)
	RefreshToken(ctx context.Context, refreshToken string) (model.Token, error)
	DeleteByID(ctx context.Context) (model.UserUpdateData, error)
	UpdateEmailRequest(ctx context.Context, newEmail string) error
	ConfirmEmailChange(ctx context.Context, token string) error
	ChangePassword(ctx context.Context, user model.User) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	UpdateUsername(ctx context.Context, username string) (model.UserUpdateData, error)
}
