package model

import "time"

type RequestChangeToken struct {
	Token     string
	UserID    int64
	NewEmail  string
	ExpiresAt time.Time
}

type EmailSendRequest struct {
	To      string
	Subject string
	Body    string
}
