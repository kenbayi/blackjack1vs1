package users

import (
	"context"
	svc "game_svc/internal/adapter/grpc/server/frontend/proto/user"
	"game_svc/internal/adapter/grpc/users/dto"
	"game_svc/internal/model"
)

type Client struct {
	client svc.UserServiceClient
}

func NewClient(client svc.UserServiceClient) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) AddBalance(ctx context.Context, request model.User) (model.User, error) {
	_, err := c.client.AddBalance(ctx, &svc.BalanceUpdateRequest{
		Id:      request.ID,
		Balance: *request.Balance,
	})

	if err != nil {
		return model.User{}, err
	}

	return model.User{}, nil
}

func (c *Client) SubtractBalance(ctx context.Context, request model.User) (model.User, error) {
	_, err := c.client.SubtractBalance(ctx, &svc.BalanceUpdateRequest{
		Id:      request.ID,
		Balance: *request.Balance,
	})

	if err != nil {
		return model.User{}, err
	}

	return model.User{}, nil
}

func (c *Client) Get(ctx context.Context, id int64) (*model.User, error) {
	resp, err := c.client.GetBalance(ctx, &svc.UserIDRequest{Id: id})
	if err != nil {
		return &model.User{}, err
	}

	return dto.FromGRPCClientGetResponse(resp), nil
}

func (c *Client) GetRating(ctx context.Context, id int64) (*model.User, error) {
	resp, err := c.client.GetRating(ctx, &svc.UserIDRequest{Id: id})
	if err != nil {
		return &model.User{}, err
	}
	return &model.User{Rating: &resp.Rating}, nil
}
