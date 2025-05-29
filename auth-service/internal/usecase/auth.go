package usecase

import (
	"auth_svc/internal/model"
	"auth_svc/pkg/def"
	"context"
	"fmt"
	"log"
	"time"
)

func (uc *User) Register(ctx context.Context, request model.User) (int64, error) {
	var customerID int64

	txFn := func(ctx context.Context) error {
		// Hash password
		hashedPassword, err := uc.passwordManager.HashPassword(request.NewPassword)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		request.PasswordHash = hashedPassword

		// Set timestamps
		now := time.Now().UTC()
		request.CreatedAt = now
		request.UpdatedAt = now

		// Create customer in database
		err = uc.repo.Create(ctx, request)
		if err != nil {
			return fmt.Errorf("failed to create customer: %w", err)
		}

		customer, err := uc.repo.GetWithFilter(ctx, model.UserFilter{Email: &request.Email})
		if err != nil {
			return fmt.Errorf("failed to get created customer: %w", err)
		}
		customerID = customer.ID

		// Publish event
		if err := uc.producer.PushCreated(ctx, customer); err != nil {
			log.Printf("failed to push customer event: %v", err)
		}

		return nil
	}

	if err := uc.callTx(ctx, txFn); err != nil {
		return 0, fmt.Errorf("registration transaction failed: %w", err)
	}

	return customerID, nil
}

func (uc *User) Login(ctx context.Context, email, password string) (model.Token, error) {
	// Get customer by email
	customer, err := uc.repo.GetWithFilter(ctx, model.UserFilter{Email: def.Pointer(email)})
	if err != nil {
		return model.Token{}, model.ErrInvalidEmail
	}

	// Verify password
	if err := uc.passwordManager.CheckPassword(customer.PasswordHash, password); err != nil {
		return model.Token{}, model.ErrInvalidPassword
	}
	// Generate tokens
	accessToken, err := uc.jwtManager.GenerateAccessToken(customer.ID, model.UserRole)
	if err != nil {
		return model.Token{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := uc.jwtManager.GenerateRefreshToken(customer.ID)
	if err != nil {
		return model.Token{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create session
	session := model.Session{
		UserID:       customer.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:    time.Now(),
	}

	if err := uc.tokenRepo.Create(ctx, session); err != nil {
		return model.Token{}, fmt.Errorf("failed to create session: %w", err)
	}

	return model.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (uc *User) RefreshToken(ctx context.Context, refreshToken string) (model.Token, error) {
	// Get existing session
	session, err := uc.tokenRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return model.Token{}, model.ErrInvalidID
	}

	// Check if refresh token is expired
	if session.ExpiresAt.Before(time.Now()) {
		return model.Token{}, model.ErrRefreshTokenExpired
	}

	// Get customer
	customer, err := uc.repo.GetWithFilter(ctx, model.UserFilter{ID: def.Pointer(session.UserID)})
	if err != nil {
		return model.Token{}, fmt.Errorf("failed to get customer: %w", err)
	}

	// Generate new tokens
	newAccessToken, err := uc.jwtManager.GenerateAccessToken(customer.ID, model.UserRole)
	if err != nil {
		return model.Token{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := uc.jwtManager.GenerateRefreshToken(customer.ID)
	if err != nil {
		return model.Token{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Delete old session and create new one
	if err := uc.tokenRepo.DeleteByToken(ctx, refreshToken); err != nil {
		return model.Token{}, fmt.Errorf("failed to delete old session: %w", err)
	}

	newSession := model.Session{
		UserID:       customer.ID,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:    time.Now(),
	}

	if err := uc.tokenRepo.Create(ctx, newSession); err != nil {
		return model.Token{}, fmt.Errorf("failed to create new session: %w", err)
	}

	return model.Token{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
