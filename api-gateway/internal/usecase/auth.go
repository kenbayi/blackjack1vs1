package usecase

import (
	"api-gateway/internal/model"
	"context"
)

type User struct {
	presenter UserPresenter
}

func NewUser(p UserPresenter) *User {
	return &User{presenter: p}
}

func (u *User) Register(ctx context.Context, user model.User) (int64, error) {
	return u.presenter.Register(ctx, user)
}

func (u *User) Login(ctx context.Context, user model.User) (model.Token, error) {
	return u.presenter.Login(ctx, user)
}

func (u *User) RefreshToken(ctx context.Context, token model.Token) (model.Token, error) {
	return u.presenter.RefreshToken(ctx, token)
}

func (u *User) DeleteByID(ctx context.Context) (model.UserUpdateData, error) {
	return u.presenter.DeleteByID(ctx)
}

func (u *User) UpdateUsername(ctx context.Context, user model.UserUpdateData) (model.UserUpdateData, error) {
	return u.presenter.UpdateUsername(ctx, user)
}

func (u *User) UpdateEmailRequest(ctx context.Context, user model.UserUpdateData) error {
	return u.presenter.UpdateEmailRequest(ctx, user)
}

func (u *User) ConfirmEmailChange(ctx context.Context, req model.RequestToChange) error {
	return u.presenter.ConfirmEmailChange(ctx, req)
}

func (u *User) ChangePassword(ctx context.Context, user model.User) error {
	return u.presenter.ChangePassword(ctx, user)
}

func (u *User) RequestPasswordReset(ctx context.Context, user model.User) error {
	return u.presenter.RequestPasswordReset(ctx, user)
}

func (u *User) ResetPassword(ctx context.Context, req model.RequestToChange) error {
	return u.presenter.ResetPassword(ctx, req)
}
