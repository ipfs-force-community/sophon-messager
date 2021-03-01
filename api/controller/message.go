package controller

import (
	"context"
	"github.com/ipfs-force-community/venus-messager/types"
)

type Message struct {
	BaseController
}

func (message Message) PushMessage(ctx context.Context, msg *types.Message) (string, error) {
	return message.Repo.MessageRepo().SaveMessage(msg)
}

func (message Message) GetMessage(ctx context.Context, uuid string) (types.Message, error) {
	return message.Repo.MessageRepo().GetMessage(uuid)
}

func (message Message) ListMessage(ctx context.Context) ([]types.Message, error) {
	return message.Repo.MessageRepo().ListMessage()
}
