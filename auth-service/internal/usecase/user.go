package usecase

import (
	"auth_svc/pkg/transactor"
	"context"
	"fmt"
	"time"

	"auth_svc/internal/model"
	"auth_svc/pkg/def"
	"auth_svc/pkg/security"
)

type User struct {
	repo            UserRepo
	tokenRepo       RefreshTokenRepo
	redisEmailRepo  RedisEmailRepo
	producer        UserEventStorage
	callTx          transactor.WithinTransactionFunc
	jwtManager      *security.JWTManager
	passwordManager *security.PasswordManager
}

func NewUser(
	repo UserRepo,
	tokenRepo RefreshTokenRepo,
	redisEmailRepo RedisEmailRepo,
	producer UserEventStorage,
	callTx transactor.WithinTransactionFunc,
	jwtManager *security.JWTManager,
	passwordManager *security.PasswordManager,
) *User {
	return &User{
		repo:            repo,
		tokenRepo:       tokenRepo,
		redisEmailRepo:  redisEmailRepo,
		producer:        producer,
		callTx:          callTx,
		jwtManager:      jwtManager,
		passwordManager: passwordManager,
	}
}

func (uc *User) DeleteByID(ctx context.Context) (model.UserUpdateData, error) {
	var updateData model.UserUpdateData

	txFn := func(ctx context.Context) error {
		token, ok := security.TokenFromCtx(ctx)
		if !ok {
			return model.ErrInvalidID
		}
		claims, err := uc.jwtManager.Verify(token)
		if err != nil {
			return model.ErrInvalidID
		}
		rawID, ok := claims["user_id"].(float64)
		if !ok {
			return model.ErrInvalidID
		}
		userID := int64(rawID)

		existing, err := uc.repo.GetWithFilter(ctx, model.UserFilter{ID: def.Pointer(userID), IsDeleted: def.Pointer(false)})
		if err != nil {
			return model.ErrInvalidID
		}

		updateData = model.UserUpdateData{
			ID:        &existing.ID,
			IsDeleted: def.Pointer(true),
			UpdatedAt: def.Pointer(time.Now()),
		}

		if err := uc.repo.PatchByID(ctx, &updateData); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		if err := uc.producer.PushDeleted(ctx, &updateData); err != nil {
			return fmt.Errorf("failed to push deleted user event: %w", err)
		}

		return nil
	}

	if err := uc.callTx(ctx, txFn); err != nil {
		return model.UserUpdateData{}, fmt.Errorf("delete user transaction failed: %w", err)
	}

	return updateData, nil
}

func (uc *User) UpdateUsername(ctx context.Context, username string) (model.UserUpdateData, error) {
	var updateData model.UserUpdateData
	txFn := func(ctx context.Context) error {
		token, ok := security.TokenFromCtx(ctx)
		println(token)
		if !ok {
			return model.ErrInvalidID
		}
		claims, err := uc.jwtManager.Verify(token)
		if err != nil {
			return model.ErrInvalidID
		}
		rawID, ok := claims["user_id"].(float64)
		if !ok {
			return model.ErrInvalidID
		}
		userID := int64(rawID)
		println(userID)

		updateData = model.UserUpdateData{
			ID:        &userID,
			Username:  &username,
			UpdatedAt: def.Pointer(time.Now()),
		}

		if err := uc.repo.PatchByID(ctx, &updateData); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		if err := uc.producer.PushUpdated(ctx, &updateData); err != nil {
			return fmt.Errorf("failed to push updated user event: %w", err)
		}
		return nil
	}
	if err := uc.callTx(ctx, txFn); err != nil {
		return model.UserUpdateData{}, fmt.Errorf("update user transaction failed: %w", err)
	}
	return updateData, nil
}
