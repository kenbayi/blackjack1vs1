package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
	"user_svc/internal/model"
	"user_svc/pkg/redis"
)

type UserCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewUserCache(client *redis.Client, ttl time.Duration) *UserCache {
	return &UserCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *UserCache) key(userID int64) string {
	return fmt.Sprintf("user:%d", userID)
}
func balanceKey(userID int64) string {
	return "user:balance:" + strconv.FormatInt(userID, 10)
}
func ratingKey(userID int64) string {
	return "user:rating:" + strconv.FormatInt(userID, 10)
}

func (c *UserCache) Get(ctx context.Context, userID int64) (model.User, error) {
	log.Printf("Got from cache")
	key := c.key(userID)
	pipe := c.client.Unwrap().Pipeline()
	getCmd := pipe.Get(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return model.User{}, err
	}
	val, err := getCmd.Result()
	if err != nil {
		return model.User{}, err
	}
	var user model.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (c *UserCache) Set(ctx context.Context, user model.User) error {
	log.Printf("Set to cache")
	key := c.key(user.ID)
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	pipe := c.client.Unwrap().Pipeline()
	setCmd := pipe.Set(ctx, key, string(data), c.ttl)

	_, err = pipe.Exec(ctx)
	return setCmd.Err()
}

func (c *UserCache) Delete(ctx context.Context, userID int64) error {
	log.Printf("Delete from cache")
	key := c.key(userID)
	pipe := c.client.Unwrap().Pipeline()
	delCmd := pipe.Del(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return delCmd.Err()
}

func (c *UserCache) GetRating(ctx context.Context, userID int64) (int64, error) {
	log.Printf("Got from cache")
	key := ratingKey(userID)
	pipe := c.client.Unwrap().Pipeline()
	getCmd := pipe.Get(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	val, err := getCmd.Result()
	if err != nil {
		return 0, err
	}
	rating, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse rating from cache: %w", err)
	}
	return rating, nil
}

func (c *UserCache) SetRating(ctx context.Context, userID int64, rating int64) error {
	log.Printf("Set to cache")
	key := ratingKey(userID)
	value := strconv.FormatInt(rating, 10)

	pipe := c.client.Unwrap().Pipeline()
	setCmd := pipe.Set(ctx, key, value, c.ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return setCmd.Err()
}

func (c *UserCache) DeleteRating(ctx context.Context, userID int64) error {
	log.Printf("Delete from cache")
	key := ratingKey(userID)
	pipe := c.client.Unwrap().Pipeline()
	delCmd := pipe.Del(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return delCmd.Err()
}

func (c *UserCache) GetBalance(ctx context.Context, userID int64) (int64, error) {
	log.Printf("Got from cache")
	key := balanceKey(userID)
	pipe := c.client.Unwrap().Pipeline()
	getCmd := pipe.Get(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	val, err := getCmd.Result()
	if err != nil {
		return 0, err
	}
	balance, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse balance from cache: %w", err)
	}
	return balance, nil
}

func (c *UserCache) SetBalance(ctx context.Context, userID int64, balance int64) error {
	log.Printf("Set to cache")
	key := balanceKey(userID)
	value := strconv.FormatInt(balance, 10)

	pipe := c.client.Unwrap().Pipeline()
	setCmd := pipe.Set(ctx, key, value, c.ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return setCmd.Err()
}
