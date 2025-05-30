package dto

import (
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"user_svc/internal/model"
)

var (
	ErrEmailAlreadyRegistered = status.Error(codes.InvalidArgument, "this email is already registered")
	ErrUserNotFound           = status.Error(codes.NotFound, "user not found")
	ErrInvalidInput           = status.Error(codes.InvalidArgument, "invalid input data")
	ErrNotEnoughBalance       = status.Error(codes.FailedPrecondition, "not enough balance")

	ErrConflict = status.Error(codes.AlreadyExists, "conflict")
)

func FromError(err error) error {
	switch {
	case errors.Is(err, model.ErrEmailAlreadyRegistered):
		return ErrEmailAlreadyRegistered
	case errors.Is(err, model.ErrUserNotFound):
		return ErrUserNotFound
	case errors.Is(err, model.ErrInvalidInput):
		return ErrInvalidInput
	case errors.Is(err, model.ErrNotEnoughBalance):
		return ErrNotEnoughBalance
	case errors.Is(err, model.ErrConflict):
		return ErrConflict

	default:
		return status.Error(codes.Internal, "something went wrong")
	}
}
