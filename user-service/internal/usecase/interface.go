package usecase

import (
	"context"
	"user_svc/internal/model"
)

type UserRepo interface {
	Create(ctx context.Context, customer model.User) error
	PatchByID(ctx context.Context, userUpdated *model.UserUpdateData) error
	GetWithFilter(ctx context.Context, filter model.UserFilter) (model.User, error)
	GetListWithFilter(ctx context.Context, filter model.UserFilter) ([]model.User, error)
	GetBalance(ctx context.Context, userID int64) (int64, error)
	UpdateBalance(ctx context.Context, userID int64, newBalance int64) error
	GetRating(ctx context.Context, userID int64) (int64, error)
	UpdateRating(ctx context.Context, userID int64, newRating int64) error
}

type UserCache interface {
	// Profile caching
	Get(ctx context.Context, userID int64) (model.User, error)
	Set(ctx context.Context, user model.User) error
	Delete(ctx context.Context, userID int64) error

	// Rating caching
	GetRating(ctx context.Context, userID int64) (int64, error)
	SetRating(ctx context.Context, userID int64, rating int64) error
	DeleteRating(ctx context.Context, userID int64) error

	// Balance caching
	GetBalance(ctx context.Context, userID int64) (int64, error)
	SetBalance(ctx context.Context, userID int64, balance int64) error
}
