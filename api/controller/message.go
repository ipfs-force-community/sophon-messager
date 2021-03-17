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
		ID:              types.NewUUID(),
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

func (message *Message) GetMessageBySignedCid(ctx context.Context, cid string) (*types.Message, error) {
	return message.MsgService.GetMessageBySignedCid(ctx, cid)
}

func (message *Message) GetMessageByUnsignedCid(ctx context.Context, cid string) (*types.Message, error) {
	return message.MsgService.GetMessageByUnsignedCid(ctx, cid)
}

func (message Message) ListMessage(ctx context.Context) ([]*types.Message, error) {
	return message.MsgService.ListMessage(ctx)
}

func (message *Message) UpdateMessageStateByCid(ctx context.Context, cid string, state types.MessageState) (string, error) {
	return message.MsgService.UpdateMessageStateByCid(ctx, cid, state)
}

func (message *Message) UpdateMessageStateByID(ctx context.Context, id types.UUID, state types.MessageState) (types.UUID, error) {
	return message.MsgService.UpdateMessageStateByID(ctx, id, state)
}

func (message *Message) UpdateAllSignedMessage(ctx context.Context) (int, error) {
	return message.MsgService.UpdateAllSignedMessage(ctx)
}

func (message *Message) UpdateSignedMessageByID(ctx context.Context, uuid types.UUID) (types.UUID, error) {
	return message.MsgService.UpdateSignedMessageByID(ctx, uuid)
}
