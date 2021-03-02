package controller

import (
	"context"

	"github.com/ipfs-force-community/venus-messager/service"
	"github.com/ipfs-force-community/venus-messager/types"
)

type Message struct {
	BaseController
	MsgService *service.MessageService
}

func (message Message) PushMessage(ctx context.Context, msg *types.Message) (string, error) {
	return message.MsgService.PushMessage(ctx, msg)
}

func (message Message) GetMessage(ctx context.Context, uuid string) (*types.Message, error) {
	return message.MsgService.GetMessage(ctx, uuid)
}

func (message Message) ListMessage(ctx context.Context) ([]*types.Message, error) {
	return message.MsgService.ListMessage(ctx)
}
