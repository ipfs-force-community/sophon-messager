package client

import (
	"context"
	"github.com/ipfs-force-community/venus-messager/types"
)

type IMessager interface {
	PushMessage(ctx context.Context, msg *types.Message) (string, error)
	GetMessage(ctx context.Context, uuid string) (types.Message, error)
	ListMessage(ctx context.Context) ([]types.Message, error)
}

var _ IMessager = (*Message)(nil)

type Message struct {
	Internal struct {
		PushMessage func(ctx context.Context, msg *types.Message) (string, error)
		GetMessage  func(ctx context.Context, uuid string) (types.Message, error)
		ListMessage func(ctx context.Context) ([]types.Message, error)
	}
}

func (message *Message) PushMessage(ctx context.Context, msg *types.Message) (string, error) {
	return message.Internal.PushMessage(ctx, msg)
}

func (message *Message) GetMessage(ctx context.Context, uuid string) (types.Message, error) {
	return message.Internal.GetMessage(ctx, uuid)
}

func (message *Message) ListMessage(ctx context.Context) ([]types.Message, error) {
	return message.Internal.ListMessage(ctx)
}
