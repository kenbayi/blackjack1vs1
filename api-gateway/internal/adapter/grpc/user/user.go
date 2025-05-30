package user

import (
	svc "api-gateway/internal/adapter/frontend/proto/user"
	"api-gateway/internal/adapter/grpc/user/dto"
	"api-gateway/internal/model"
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
)

type User struct {
	user svc.UserServiceClient
}

func NewUser(client svc.UserServiceClient) *User {
	return &User{user: client}
}

func (s *User) GetBalance(ctx context.Context, request model.UserProfile) (model.UserProfile, error) {
	resp, err := s.user.GetBalance(ctx, &svc.UserIDRequest{
		Id: request.ID,
	})
	if err != nil {
		return model.UserProfile{}, err
	}
	return *dto.FromGRPCGetBalanceResponse(resp), nil
}

func (s *User) AddBalance(ctx context.Context, request model.UserProfile) (*emptypb.Empty, error) {
	_, err := s.user.AddBalance(ctx, &svc.BalanceUpdateRequest{
		Id:      request.ID,
		Balance: *request.Balance,
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *User) SubtractBalance(ctx context.Context, request model.UserProfile) (*emptypb.Empty, error) {
	_, err := s.user.SubtractBalance(ctx, &svc.BalanceUpdateRequest{
		Id:      request.ID,
		Balance: *request.Balance,
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *User) GetProfile(ctx context.Context, request model.UserProfile) (model.UserProfile, error) {
	resp, err := s.user.GetProfile(ctx, &svc.UserIDRequest{
		Id: request.ID,
	})
	log.Printf("GetProfile Presenter: after request: %v", resp)
	if err != nil {
		return model.UserProfile{}, err
	}
	log.Printf("GetProfile Presenter: after dto: %v", *dto.FromGRPCGetProfileResponse(resp))

	return *dto.FromGRPCGetProfileResponse(resp), nil
}

func (s *User) UpdateProfile(ctx context.Context, request model.UserProfile) (*emptypb.Empty, error) {
	// 1. Create the gRPC request struct first.
	grpcRequest := &svc.UpdateProfileRequest{
		Id: request.ID,
	}

	// 2. Only set fields if the corresponding pointers in the model are not nil.
	if request.Nickname != nil {
		grpcRequest.Nickname = *request.Nickname
	}

	if request.Bio != nil {
		grpcRequest.Bio = *request.Bio
	}

	// 3. Make the gRPC call with the safely constructed request.
	_, err := s.user.UpdateProfile(ctx, grpcRequest)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *User) GetRating(ctx context.Context, request model.UserProfile) (model.UserProfile, error) {
	resp, err := s.user.GetRating(ctx, &svc.UserIDRequest{
		Id: request.ID,
	})
	if err != nil {
		return model.UserProfile{}, err
	}
	return *dto.FromGRPCGetRatingResponse(resp), nil
}

func (s *User) UpdateRating(ctx context.Context, request model.UserProfile) (*emptypb.Empty, error) {
	_, err := s.user.UpdateRating(ctx, &svc.RatingUpdateResponse{
		Id:     request.ID,
		Rating: *request.Rating,
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}
