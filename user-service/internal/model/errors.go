package model

import "errors"

var (
	ErrNotFound               = errors.New("not found")
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrInvalidInput           = errors.New("invalid input data")
	ErrNotEnoughBalance       = errors.New("not enough balance")
	ErrConflict               = errors.New("conflict")
	ErrUserNotFound           = errors.New("user not found")
)
