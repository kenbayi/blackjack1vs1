package usecase

import (
	"api-gateway/internal/model"
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserPresenter interface {
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

type StatisticsPresenter interface {
	GetGeneralGameStats(ctx context.Context) (*model.GeneralGameStats, error)
	GetUserGameStats(ctx context.Context, userID int64) (*model.UserGameStats, error)
	GetLeaderboard(ctx context.Context, req model.Leaderboard) (*model.Leaderboard, error)
}

type UserProfilePresenter interface {
	GetBalance(ctx context.Context, request model.UserProfile) (model.UserProfile, error)
	AddBalance(ctx context.Context, request model.UserProfile) (*emptypb.Empty, error)
	SubtractBalance(ctx context.Context, request model.UserProfile) (*emptypb.Empty, error)
	GetProfile(ctx context.Context, request model.UserProfile) (model.UserProfile, error)
	UpdateProfile(ctx context.Context, request model.UserProfile) (*emptypb.Empty, error)
	GetRating(ctx context.Context, request model.UserProfile) (model.UserProfile, error)
	UpdateRating(ctx context.Context, request model.UserProfile) (*emptypb.Empty, error)
}
