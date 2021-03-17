package client

import (
	"context"
	"time"

	"github.com/filecoin-project/go-address"
	venusTypes "github.com/filecoin-project/venus/pkg/types"
	"golang.org/x/xerrors"

	"github.com/ipfs-force-community/venus-messager/types"
)

type IMessager interface {
	WaitMessage(ctx context.Context, uuid types.UUID, confidence uint64) (*types.Message, error)
	PushMessage(ctx context.Context, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (types.UUID, error)
	PushMessageWithId(ctx context.Context, uuid types.UUID, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (types.UUID, error)
	GetMessage(ctx context.Context, uuid types.UUID) (*types.Message, error)
	GetMessageBySignedCid(ctx context.Context, cid string) (*types.Message, error)
	GetMessageByUnsignedCid(ctx context.Context, cid string) (*types.Message, error)
	ListMessage(ctx context.Context) ([]*types.Message, error)
	UpdateMessageStateByCid(ctx context.Context, cid string, state types.MessageState) (string, error)
	UpdateMessageStateByID(ctx context.Context, id types.UUID, state types.MessageState) (types.UUID, error)
	UpdateAllSignedMessage(ctx context.Context) (int, error)
	UpdateSignedMessageByID(ctx context.Context, uuid types.UUID) (types.UUID, error)

	SaveWallet(ctx context.Context, wallet *types.Wallet) (types.UUID, error)
	GetWalletByID(ctx context.Context, uuid types.UUID) (*types.Wallet, error)
	GetWalletByName(ctx context.Context, name string) (*types.Wallet, error)
	ListWallet(ctx context.Context) ([]*types.Wallet, error)
	ListRemoteWalletAddress(ctx context.Context, uuid types.UUID) ([]address.Address, error)

	SaveAddress(ctx context.Context, address *types.Address) (string, error)
	GetAddress(ctx context.Context, addr string) (*types.Address, error)
	HasAddress(ctx context.Context, addr address.Address) (bool, error)
	ListAddress(ctx context.Context) ([]*types.Address, error)
	UpdateNonce(ctx context.Context, uuid types.UUID, nonce uint64) (types.UUID, error)
}

var _ IMessager = (*Message)(nil)

type Message struct {
	Internal struct {
		WaitMessage             func(ctx context.Context, uuid types.UUID, confidence uint64) (*types.Message, error)
		PushMessage             func(ctx context.Context, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (types.UUID, error)
		PushMessageWithId       func(ctx context.Context, uuid types.UUID, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (types.UUID, error)
		GetMessage              func(ctx context.Context, uuid types.UUID) (*types.Message, error)
		GetMessageBySignedCid   func(ctx context.Context, cid string) (*types.Message, error)
		GetMessageByUnsignedCid func(ctx context.Context, cid string) (*types.Message, error)
		ListMessage             func(ctx context.Context) ([]*types.Message, error)
		UpdateMessageStateByCid func(ctx context.Context, cid string, state types.MessageState) (string, error)
		UpdateMessageStateByID  func(ctx context.Context, id types.UUID, state types.MessageState) (types.UUID, error)
		UpdateAllSignedMessage  func(ctx context.Context) (int, error)
		UpdateSignedMessageByID func(ctx context.Context, uuid types.UUID) (types.UUID, error)

		SaveWallet              func(ctx context.Context, wallet *types.Wallet) (types.UUID, error)
		GetWalletByID           func(ctx context.Context, uuid types.UUID) (*types.Wallet, error)
		GetWalletByName         func(ctx context.Context, name string) (*types.Wallet, error)
		ListWallet              func(ctx context.Context) ([]*types.Wallet, error)
		ListRemoteWalletAddress func(ctx context.Context, uuid types.UUID) ([]address.Address, error)

		SaveAddress func(ctx context.Context, address *types.Address) (string, error)
		GetAddress  func(ctx context.Context, addr string) (*types.Address, error)
		HasAddress  func(ctx context.Context, addr address.Address) (bool, error)
		ListAddress func(ctx context.Context) ([]*types.Address, error)
		UpdateNonce func(ctx context.Context, uuid types.UUID, nonce uint64) (types.UUID, error)
	}
}

func (message *Message) PushMessage(ctx context.Context, msg *venusTypes.Message, meta *types.MsgMeta) (types.UUID, error) {
	return message.Internal.PushMessage(ctx, msg, meta)
}

func (message *Message) PushMessageWithId(ctx context.Context, uuid types.UUID, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (types.UUID, error) {
	return message.Internal.PushMessageWithId(ctx, uuid, msg, meta)
}

func (message *Message) GetMessage(ctx context.Context, uuid types.UUID) (*types.Message, error) {
	return message.Internal.GetMessage(ctx, uuid)
}

func (message *Message) GetMessageByUnsignedCid(ctx context.Context, cid string) (*types.Message, error) {
	return message.Internal.GetMessageByUnsignedCid(ctx, cid)
}

func (message *Message) GetMessageBySignedCid(ctx context.Context, cid string) (*types.Message, error) {
	return message.Internal.GetMessageBySignedCid(ctx, cid)
}

func (message *Message) ListMessage(ctx context.Context) ([]*types.Message, error) {
	return message.Internal.ListMessage(ctx)
}

func (message *Message) UpdateMessageStateByCid(ctx context.Context, cid string, state types.MessageState) (string, error) {
	return message.Internal.UpdateMessageStateByCid(ctx, cid, state)
}

func (message *Message) UpdateMessageStateByID(ctx context.Context, id types.UUID, state types.MessageState) (types.UUID, error) {
	return message.Internal.UpdateMessageStateByID(ctx, id, state)
}

func (message *Message) UpdateAllSignedMessage(ctx context.Context) (int, error) {
	return message.Internal.UpdateAllSignedMessage(ctx)
}

func (message *Message) UpdateSignedMessageByID(ctx context.Context, uuid types.UUID) (types.UUID, error) {
	return message.Internal.UpdateSignedMessageByID(ctx, uuid)
}

func (message *Message) SaveWallet(ctx context.Context, wallet *types.Wallet) (types.UUID, error) {
	return message.Internal.SaveWallet(ctx, wallet)
}

func (message *Message) GetWalletByID(ctx context.Context, uuid types.UUID) (*types.Wallet, error) {
	return message.Internal.GetWalletByID(ctx, uuid)
}

func (message *Message) GetWalletByName(ctx context.Context, name string) (*types.Wallet, error) {
	return message.Internal.GetWalletByName(ctx, name)
}

func (message *Message) ListRemoteWalletAddress(ctx context.Context, uuid types.UUID) ([]address.Address, error) {
	return message.Internal.ListRemoteWalletAddress(ctx, uuid)
}

func (message *Message) ListWallet(ctx context.Context) ([]*types.Wallet, error) {
	return message.Internal.ListWallet(ctx)
}

func (message *Message) SaveAddress(ctx context.Context, address *types.Address) (string, error) {
	return message.Internal.SaveAddress(ctx, address)
}

func (message *Message) GetAddress(ctx context.Context, addr string) (*types.Address, error) {
	return message.Internal.GetAddress(ctx, addr)
}

func (message *Message) HasAddress(ctx context.Context, addr address.Address) (bool, error) {
	return message.Internal.HasAddress(ctx, addr)
}

func (message *Message) ListAddress(ctx context.Context) ([]*types.Address, error) {
	return message.Internal.ListAddress(ctx)
}

func (message *Message) UpdateNonce(ctx context.Context, uuid types.UUID, nonce uint64) (types.UUID, error) {
	return message.Internal.UpdateNonce(ctx, uuid, nonce)
}

func (message *Message) WaitMessage(ctx context.Context, uuid types.UUID, confidence uint64) (*types.Message, error) {
	tm := time.NewTicker(time.Second * 30)
	defer tm.Stop()

	for {
		select {
		case <-tm.C:
			msg, err := message.Internal.GetMessage(ctx, uuid)
			if err != nil {
				return nil, err
			}

			if msg.State == types.OnChainMsg && msg.Confidence > int64(confidence) {
				return msg, nil
			}
			continue
		case <-ctx.Done():
			return nil, xerrors.New("exit by client ")
		}
	}
}
