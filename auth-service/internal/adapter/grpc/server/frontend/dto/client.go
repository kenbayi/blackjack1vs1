package dto

import (
	svc "auth_svc/internal/adapter/grpc/server/frontend/proto/user"
	"auth_svc/internal/model"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func ToUserFromRegisterRequest(req *svc.RegisterRequest) (model.User, error) {
	return model.User{
		Username:    req.Username,
		Email:       req.Email,
		NewPassword: req.Password,
	}, nil
}

func ToProtoTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func ToUserFromChangePasswordRequest(req *svc.ChangePasswordRequest) (model.User, error) {
	return model.User{
		ID:              req.Id,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}, nil
}
