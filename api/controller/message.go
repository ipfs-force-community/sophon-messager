package controller

import (
	"context"
	"github.com/filecoin-project/go-address"

	venusTypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs-force-community/venus-messager/service"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/ipfs/go-cid"
)

type Message struct {
	BaseController
	MsgService *service.MessageService
}

func (message Message) PushMessage(ctx context.Context, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (types.UUID, error) {
	newId := types.NewUUID()
	err := message.MsgService.PushMessage(ctx, &types.Message{
		ID:              newId,
		UnsignedMessage: *msg,
		Meta:            meta,
		State:           types.UnFillMsg,
	})
	if err != nil {
		return types.UUID{}, nil
	}
	return newId, nil
}

func (message Message) PushMessageWithId(ctx context.Context, uuid types.UUID, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (types.UUID, error) {
	return uuid, message.MsgService.PushMessage(ctx, &types.Message{
		ID:              uuid,
		UnsignedMessage: *msg,
		Meta:            meta,
		State:           types.UnFillMsg,
	})
}

func (message Message) GetMessageByUid(ctx context.Context, uuid types.UUID) (*types.Message, error) {
	return message.MsgService.GetMessageByUid(ctx, uuid)
}

func (message Message) GetMessageByCid(ctx context.Context, id cid.Cid) (*types.Message, error) {
	return message.MsgService.GetMessageByCid(ctx, id)
}

func (message Message) GetMessageState(ctx context.Context, uuid types.UUID) (types.MessageState, error) {
	return message.MsgService.GetMessageState(ctx, uuid)
}

func (message Message) GetMessageBySignedCid(ctx context.Context, cid cid.Cid) (*types.Message, error) {
	return message.MsgService.GetMessageBySignedCid(ctx, cid)
}

func (message Message) GetMessageByUnsignedCid(ctx context.Context, cid cid.Cid) (*types.Message, error) {
	return message.MsgService.GetMessageByUnsignedCid(ctx, cid)
}

func (message Message) GetMessageByFromAndNonce(ctx context.Context, from address.Address, nonce uint64) (*types.Message, error) {
	return message.MsgService.GetMessageByFromAndNonce(ctx, from, nonce)
}

func (message Message) ListMessage(ctx context.Context) ([]*types.Message, error) {
	return message.MsgService.ListMessage(ctx)
}

func (message Message) UpdateMessageStateByCid(ctx context.Context, cid string, state types.MessageState) (string, error) {
	return message.MsgService.UpdateMessageStateByCid(ctx, cid, state)
}

func (message Message) UpdateMessageStateByID(ctx context.Context, id types.UUID, state types.MessageState) (types.UUID, error) {
	return message.MsgService.UpdateMessageStateByID(ctx, id, state)
}

func (message Message) UpdateAllFilledMessage(ctx context.Context) (int, error) {
	return message.MsgService.UpdateAllFilledMessage(ctx)
}

func (message Message) UpdateFilledMessageByID(ctx context.Context, uuid types.UUID) (types.UUID, error) {
	return message.MsgService.UpdateSignedMessageByID(ctx, uuid)
}

func (message Message) ReplaceMessage(ctx context.Context, uuid types.UUID, auto bool, maxFee string, gasLimit int64, gasPremium string, gasFeecap string) (cid.Cid, error) {
	return message.MsgService.ReplaceMessage(ctx, uuid, auto, maxFee, gasLimit, gasPremium, gasFeecap)
}
