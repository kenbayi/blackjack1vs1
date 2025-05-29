package dto

import (
	events "email_svc/internal/adapter/proto"
	"email_svc/internal/model"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
)

func ToEmailDetail(msg *nats.Msg) (model.EmailSentDetail, error) {
	var pbDetail events.EmailSendRequest
	err := proto.Unmarshal(msg.Data, &pbDetail)
	if err != nil {
		return model.EmailSentDetail{}, fmt.Errorf("proto.Unmarshall: %w", err)
	}

	return model.EmailSentDetail{
		To:      pbDetail.To,
		Subject: pbDetail.Subject,
		Body:    pbDetail.Body,
	}, nil
}
