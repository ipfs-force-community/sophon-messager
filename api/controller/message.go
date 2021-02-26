package controller

import (
	"context"
	"github.com/ipfs-force-community/venus-messager/models"
	"time"
)

type Message struct {
	BaseController
}

func (message Message) PushMessage(ctx context.Context, uuid string) (string, error) {
	msgRepo := models.NewMessageRepo(message.Repo)
	msg := &models.Message{
		Id:        uuid,
		Version:   0,
		To:        "",
		From:      "",
		Nonce:     0,
		GasLimit:  0,
		Method:    0,
		Params:    nil,
		SignData:  nil,
		IsDeleted: 0,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
	return msgRepo.SaveMessage(msg)
}
