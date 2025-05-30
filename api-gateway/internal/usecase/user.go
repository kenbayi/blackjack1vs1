package usecase

import (
	"api-gateway/internal/model"
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserProfile struct {
	presenter UserProfilePresenter
}

func NewUserProfile(p UserProfilePresenter) *UserProfile {
	return &UserProfile{presenter: p}
}

func (u *UserProfile) GetBalance(ctx context.Context, request model.UserProfile) (model.UserProfile, error) {
	return u.presenter.GetBalance(ctx, request)
}

func (u *UserProfile) AddBalance(ctx context.Context, request model.UserProfile) (*emptypb.Empty, error) {
	return u.presenter.AddBalance(ctx, request)
}

func (u *UserProfile) SubtractBalance(ctx context.Context, request model.UserProfile) (*emptypb.Empty, error) {
	return u.presenter.SubtractBalance(ctx, request)
}

func (u *UserProfile) GetProfile(ctx context.Context, request model.UserProfile) (model.UserProfile, error) {
	return u.presenter.GetProfile(ctx, request)
}

func (u *UserProfile) UpdateProfile(ctx context.Context, request model.UserProfile) (*emptypb.Empty, error) {
	return u.presenter.UpdateProfile(ctx, request)
}

func (u *UserProfile) GetRating(ctx context.Context, request model.UserProfile) (model.UserProfile, error) {
	return u.presenter.GetRating(ctx, request)
}

func (u *UserProfile) UpdateRating(ctx context.Context, request model.UserProfile) (*emptypb.Empty, error) {
	return u.presenter.UpdateRating(ctx, request)
}
