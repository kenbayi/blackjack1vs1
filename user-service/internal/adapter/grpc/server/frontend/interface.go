package frontend

import (
	"context"
	"user_svc/internal/model"
)

type UserUsecase interface {
	DeleteByID(ctx context.Context, request model.UserUpdateData) error
	UpdateUsername(ctx context.Context, updateData model.UserUpdateData) error
	RecordUser(ctx context.Context, request model.User) error
	GetBalance(ctx context.Context, userID int64) (int64, error)
	AddBalance(ctx context.Context, userID int64, delta int64) error
	SubtractBalance(ctx context.Context, userID int64, delta int64) error
	GetRating(ctx context.Context, userID int64) (int64, error)
	UpdateRating(ctx context.Context, userID int64, newRating int64) error
	GetProfile(ctx context.Context, userID int64) (model.User, error)
	UpdateProfile(ctx context.Context, update model.UserUpdateData) error
}
