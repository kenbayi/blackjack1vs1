package usecase

import (
	"auth_svc/internal/model"
	"context"
)

type UserRepo interface {
	Create(ctx context.Context, customer model.User) error
	PatchByID(ctx context.Context, userUpdated *model.UserUpdateData) error
	GetWithFilter(ctx context.Context, filter model.UserFilter) (model.User, error)
	GetListWithFilter(ctx context.Context, filter model.UserFilter) ([]model.User, error)
}

type RefreshTokenRepo interface {
	Create(ctx context.Context, session model.Session) error
	GetByToken(ctx context.Context, token string) (model.Session, error)
	DeleteByToken(ctx context.Context, token string) error
}

type UserEventStorage interface {
	PushCreated(ctx context.Context, client model.User) error
	PushUpdated(ctx context.Context, client *model.UserUpdateData) error
	PushDeleted(ctx context.Context, client *model.UserUpdateData) error
	PushEmailChangeRequest(ctx context.Context, req model.EmailSendRequest) error
}

type RedisEmailRepo interface {
	Save(ctx context.Context, token model.RequestChangeToken) error
	Get(ctx context.Context, token string) (model.RequestChangeToken, error)
	Delete(ctx context.Context, token string) error
}
