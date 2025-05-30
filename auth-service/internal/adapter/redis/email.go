package redis

import (
	"auth_svc/pkg/redis"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"auth_svc/internal/model"
)

type Email struct {
	client *redis.Client
	ttl    time.Duration
}

func NewEmail(client *redis.Client, ttl time.Duration) *Email {
	return &Email{
		client: client,
		ttl:    ttl,
	}
}

func (e *Email) Save(ctx context.Context, token model.RequestChangeToken) error {
	pipe := e.client.Unwrap().Pipeline()
	key := fmt.Sprintf("email_token:%s", token.Token)

	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	pipe.Set(ctx, key, data, e.ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save token to redis: %w", err)
	}

	return nil
}

func (e *Email) Get(ctx context.Context, token string) (model.RequestChangeToken, error) {
	key := fmt.Sprintf("email_token:%s", token)

	val, err := e.client.Unwrap().Get(ctx, key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return model.RequestChangeToken{}, fmt.Errorf("token not found or expired")
		}
		return model.RequestChangeToken{}, fmt.Errorf("failed to get token from redis: %w", err)
	}

	var emailToken model.RequestChangeToken
	err = json.Unmarshal([]byte(val), &emailToken)
	if err != nil {
		return model.RequestChangeToken{}, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return emailToken, nil
}

func (e *Email) Delete(ctx context.Context, token string) error {
	key := fmt.Sprintf("email_token:%s", token)

	pipe := e.client.Unwrap().Pipeline()
	pipe.Del(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete token from redis: %w", err)
	}

	return nil
}
