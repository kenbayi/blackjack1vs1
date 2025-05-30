package usecase

import (
	"context"
	"errors"
	"fmt"
	"user_svc/pkg/transactor"

	"user_svc/internal/model"
)

type User struct {
	repo   UserRepo
	callTx transactor.WithinTransactionFunc
	cache  UserCache
}

func NewUser(
	repo UserRepo,
	callTx transactor.WithinTransactionFunc,
	cache UserCache,
) *User {
	return &User{
		repo:   repo,
		callTx: callTx,
		cache:  cache,
	}
}

func (uc *User) DeleteByID(ctx context.Context, request model.UserUpdateData) error {
	if err := uc.repo.PatchByID(ctx, &request); err != nil {
		// Если пользователь не найден, лучше возвращать конкретную ошибку из модели
		if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrUserNotFound) {
			return model.ErrUserNotFound
		}
		return err
	}
	return uc.cache.Delete(ctx, *request.ID)
}

func (uc *User) UpdateUsername(ctx context.Context, updateData model.UserUpdateData) error {
	if err := uc.repo.PatchByID(ctx, &updateData); err != nil {
		if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrUserNotFound) {
			return model.ErrUserNotFound
		}
		return err
	}
	return uc.cache.Delete(ctx, *updateData.ID)
}

func (uc *User) RecordUser(ctx context.Context, request model.User) error {
	err := uc.repo.Create(ctx, request)
	if err != nil {
		if errors.Is(err, model.ErrEmailAlreadyRegistered) {
			return model.ErrEmailAlreadyRegistered
		}
		return err
	}

	return uc.cache.Set(ctx, request)
}

func (uc *User) GetBalance(ctx context.Context, userID int64) (int64, error) {
	balance, err := uc.cache.GetBalance(ctx, userID)
	if err == nil {
		return balance, nil
	}
	balance, err = uc.repo.GetBalance(ctx, userID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrUserNotFound) {
			return 0, model.ErrUserNotFound
		}
		return 0, err
	}

	_ = uc.cache.SetBalance(ctx, userID, balance)
	return balance, nil
}

func (uc *User) AddBalance(ctx context.Context, userID int64, delta int64) error {
	balance, err := uc.GetBalance(ctx, userID)
	if err != nil {
		return err
	}
	newBalance := balance + delta

	if err := uc.repo.UpdateBalance(ctx, userID, newBalance); err != nil {
		return err
	}

	return uc.cache.SetBalance(ctx, userID, newBalance)
}

func (uc *User) SubtractBalance(ctx context.Context, userID int64, delta int64) error {
	balance, err := uc.GetBalance(ctx, userID)
	if err != nil {
		return err
	}
	if balance < delta {
		return model.ErrNotEnoughBalance
	}
	newBalance := balance - delta

	if err := uc.repo.UpdateBalance(ctx, userID, newBalance); err != nil {
		return err
	}

	return uc.cache.SetBalance(ctx, userID, newBalance)
}

func (uc *User) GetRating(ctx context.Context, userID int64) (int64, error) {
	if rating, err := uc.cache.GetRating(ctx, userID); err == nil {
		return rating, nil
	}
	rating, err := uc.repo.GetRating(ctx, userID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrUserNotFound) {
			return 0, model.ErrUserNotFound
		}
		return 0, err
	}
	_ = uc.cache.SetRating(ctx, userID, rating)
	return rating, nil
}

func (uc *User) UpdateRating(ctx context.Context, userID int64, newRating int64) error {
	if err := uc.repo.UpdateRating(ctx, userID, newRating); err != nil {
		if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrUserNotFound) {
			return model.ErrUserNotFound
		}
		return err
	}
	return uc.cache.DeleteRating(ctx, userID)
}

func (uc *User) GetProfile(ctx context.Context, userID int64) (model.User, error) {
	user, err := uc.cache.Get(ctx, userID)
	if err == nil {
		return user, nil
	}

	user, err = uc.repo.GetWithFilter(ctx, model.UserFilter{ID: &userID})
	if err != nil {
		if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrUserNotFound) {
			return model.User{}, model.ErrUserNotFound
		}
		return model.User{}, err
	}

	_ = uc.cache.Set(ctx, user)
	return user, nil
}

func (uc *User) UpdateProfile(ctx context.Context, update model.UserUpdateData) error {
	txFn := func(ctx context.Context) error {
		if err := uc.repo.PatchByID(ctx, &update); err != nil {
			if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrUserNotFound) {
				return model.ErrUserNotFound
			}
			return err
		}
		return uc.cache.Delete(ctx, *update.ID)
	}
	if err := uc.callTx(ctx, txFn); err != nil {
		return fmt.Errorf("update user transaction failed: %w", err)
	}
	return uc.cache.Delete(ctx, *update.ID)
}
