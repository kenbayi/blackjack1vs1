package dto

import (
	svc "api-gateway/internal/adapter/frontend/proto/auth"
	"api-gateway/internal/model"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func FromGRPCClientCreateResponse(resp *svc.LoginResponse) model.Token {
	return model.Token{RefreshToken: resp.RefreshToken, AccessToken: resp.AccessToken}
}

func FromGRPCClientRegisterResponse(resp *svc.RegisterResponse) int64 {
	return resp.Id
}

func FromGRPCClientRefreshTokenResponse(resp *svc.RefreshTokenResponse) model.Token {
	return model.Token{RefreshToken: resp.RefreshToken, AccessToken: resp.AccessToken}
}
func FromGRPCClientDeleteResponse(resp *svc.DeleteByIDResponse) model.UserUpdateData {
	return model.UserUpdateData{
		ID:        &resp.Id,
		UpdatedAt: ProtoTimestampToTimePtr(resp.UpdatedAt),
	}
}

func FromGRPCClientUpdateUsernameResponse(resp *svc.UpdateUsernameResponse) model.UserUpdateData {
	return model.UserUpdateData{
		ID:        &resp.Id,
		Username:  &resp.Username,
		UpdatedAt: ProtoTimestampToTimePtr(resp.UpdatedAt),
	}
}

func ProtoTimestampToTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}
