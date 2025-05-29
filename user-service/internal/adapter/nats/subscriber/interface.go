package subscriber

import (
	"context"
	"user_svc/internal/model"
)

type UserUsecase interface {
	DeleteByID(ctx context.Context, request model.UserUpdateData) error
	UpdateUsername(ctx context.Context, updateData model.UserUpdateData) error
	RecordUser(ctx context.Context, request model.User) error
}
