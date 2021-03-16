package controller

import (
	"context"
	venusTypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs-force-community/venus-messager/service"
	"github.com/ipfs-force-community/venus-messager/types"
)

type Message struct {
	BaseController
	MsgService *service.MessageService
}

func (message Message) PushMessage(ctx context.Context, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (types.UUID, error) {
	return message.MsgService.PushMessage(ctx, &types.Message{
		UnsignedMessage: *msg,
		Meta:            meta,
		State:           types.UnFillMsg,
	})
}

func (message Message) PushMessageWithId(ctx context.Context, uuid types.UUID, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (types.UUID, error) {
	return message.MsgService.PushMessage(ctx, &types.Message{
		ID:              uuid,
		UnsignedMessage: *msg,
		Meta:            meta,
		State:           types.UnFillMsg,
	})
}
func (message Message) GetMessage(ctx context.Context, uuid types.UUID) (*types.Message, error) {
	return message.MsgService.GetMessage(ctx, uuid)
}

func (message Message) GetMessageState(ctx context.Context, uuid types.UUID) (types.MessageState, error) {
	return message.MsgService.GetMessageState(ctx, uuid)
}

func (message Message) ListMessage(ctx context.Context) ([]*types.Message, error) {
	return message.MsgService.ListMessage(ctx)
}
