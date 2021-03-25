package client

import (
	"context"
	"time"

	"github.com/filecoin-project/go-address"
	venusTypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs/go-cid"
	"golang.org/x/xerrors"

	"github.com/ipfs-force-community/venus-messager/types"
)

type IMessager interface {
	WaitMessage(ctx context.Context, id string, confidence uint64) (*types.Message, error)
	PushMessage(ctx context.Context, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (string, error)
	PushMessageWithId(ctx context.Context, id string, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (string, error)
	GetMessageByUid(ctx context.Context, id string) (*types.Message, error)
	GetMessageByCid(ctx context.Context, id cid.Cid) (*types.Message, error)
	GetMessageBySignedCid(ctx context.Context, cid cid.Cid) (*types.Message, error)
	GetMessageByUnsignedCid(ctx context.Context, cid cid.Cid) (*types.Message, error)
	GetMessageByFromAndNonce(ctx context.Context, from address.Address, nonce uint64) (*types.Message, error)
	ListMessage(ctx context.Context) ([]*types.Message, error)
	UpdateMessageStateByCid(ctx context.Context, cid cid.Cid, state types.MessageState) (cid.Cid, error)
	UpdateMessageStateByID(ctx context.Context, id string, state types.MessageState) (string, error)
	UpdateAllFilledMessage(ctx context.Context) (int, error)
	UpdateFilledMessageByID(ctx context.Context, id string) (string, error)
	ReplaceMessage(ctx context.Context, id string, auto bool, maxFee string, gasLimit int64, gasPremium string, gasFeecap string) (cid.Cid, error)

	SaveWallet(ctx context.Context, wallet *types.Wallet) (types.UUID, error)
	GetWalletByID(ctx context.Context, uuid types.UUID) (*types.Wallet, error)
	GetWalletByName(ctx context.Context, name string) (*types.Wallet, error)
	HasWallet(Context context.Context, name string) (bool, error)
	ListWallet(ctx context.Context) ([]*types.Wallet, error)
	ListRemoteWalletAddress(ctx context.Context, uuid types.UUID) ([]address.Address, error)
	DeleteWallet(ctx context.Context, name string) (string, error)
	UpdateWallet(ctx context.Context, wallet *types.Wallet) (string, error)

	SaveAddress(ctx context.Context, address *types.Address) (string, error)
	GetAddress(ctx context.Context, addr address.Address) (*types.Address, error)
	HasAddress(ctx context.Context, addr address.Address) (bool, error)
	ListAddress(ctx context.Context) ([]*types.Address, error)
	UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) (address.Address, error)
	DeleteAddress(ctx context.Context, addr address.Address) (address.Address, error)
	ForbiddenAddress(ctx context.Context, addr address.Address) (address.Address, error)
	ActiveAddress(ctx context.Context, addr address.Address) (address.Address, error)
	UpdateSelectMsgNum(ctx context.Context, addr address.Address, num int) (address.Address, error)
}

var _ IMessager = (*Message)(nil)

type Message struct {
	Internal struct {
		WaitMessage              func(ctx context.Context, id string, confidence uint64) (*types.Message, error)
		PushMessage              func(ctx context.Context, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (string, error)
		PushMessageWithId        func(ctx context.Context, id string, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (string, error)
		GetMessageByUid          func(ctx context.Context, id string) (*types.Message, error)
		GetMessageByCid          func(ctx context.Context, id cid.Cid) (*types.Message, error)
		GetMessageBySignedCid    func(ctx context.Context, cid cid.Cid) (*types.Message, error)
		GetMessageByUnsignedCid  func(ctx context.Context, cid cid.Cid) (*types.Message, error)
		GetMessageByFromAndNonce func(ctx context.Context, from address.Address, nonce uint64) (*types.Message, error)
		ListMessage              func(ctx context.Context) ([]*types.Message, error)
		UpdateMessageStateByCid  func(ctx context.Context, cid cid.Cid, state types.MessageState) (cid.Cid, error)
		UpdateMessageStateByID   func(ctx context.Context, id string, state types.MessageState) (string, error)
		UpdateAllFilledMessage   func(ctx context.Context) (int, error)
		UpdateFilledMessageByID  func(ctx context.Context, id string) (string, error)
		ReplaceMessage           func(ctx context.Context, id string, auto bool, maxFee string, gasLimit int64, gasPremium string, gasFeecap string) (cid.Cid, error)

		SaveWallet              func(ctx context.Context, wallet *types.Wallet) (types.UUID, error)
		GetWalletByID           func(ctx context.Context, uuid types.UUID) (*types.Wallet, error)
		GetWalletByName         func(ctx context.Context, name string) (*types.Wallet, error)
		HasWallet               func(ctx context.Context, name string) (bool, error)
		ListWallet              func(ctx context.Context) ([]*types.Wallet, error)
		ListRemoteWalletAddress func(ctx context.Context, uuid types.UUID) ([]address.Address, error)
		DeleteWallet            func(ctx context.Context, name string) (string, error)
		UpdateWallet            func(ctx context.Context, wallet *types.Wallet) (string, error)

		SaveAddress        func(ctx context.Context, address *types.Address) (string, error)
		GetAddress         func(ctx context.Context, addr address.Address) (*types.Address, error)
		HasAddress         func(ctx context.Context, addr address.Address) (bool, error)
		ListAddress        func(ctx context.Context) ([]*types.Address, error)
		UpdateNonce        func(ctx context.Context, addr address.Address, nonce uint64) (address.Address, error)
		DeleteAddress      func(ctx context.Context, addr address.Address) (address.Address, error)
		ForbiddenAddress   func(ctx context.Context, addr address.Address) (address.Address, error)
		ActiveAddress      func(ctx context.Context, addr address.Address) (address.Address, error)
		UpdateSelectMsgNum func(ctx context.Context, addr address.Address, num int) (address.Address, error)
	}
}

func (message *Message) PushMessage(ctx context.Context, msg *venusTypes.Message, meta *types.MsgMeta) (string, error) {
	return message.Internal.PushMessage(ctx, msg, meta)
}

func (message *Message) PushMessageWithId(ctx context.Context, id string, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (string, error) {
	return message.Internal.PushMessageWithId(ctx, id, msg, meta)
}

func (message *Message) GetMessageByUid(ctx context.Context, id string) (*types.Message, error) {
	return message.Internal.GetMessageByUid(ctx, id)
}

func (message *Message) GetMessageByCid(ctx context.Context, id cid.Cid) (*types.Message, error) {
	return message.Internal.GetMessageByCid(ctx, id)
}

func (message *Message) GetMessageByUnsignedCid(ctx context.Context, cid cid.Cid) (*types.Message, error) {
	return message.Internal.GetMessageByUnsignedCid(ctx, cid)
}

func (message *Message) GetMessageBySignedCid(ctx context.Context, cid cid.Cid) (*types.Message, error) {
	return message.Internal.GetMessageBySignedCid(ctx, cid)
}

func (message *Message) GetMessageByFromAndNonce(ctx context.Context, from address.Address, nonce uint64) (*types.Message, error) {
	return message.Internal.GetMessageByFromAndNonce(ctx, from, nonce)
}

func (message *Message) ListMessage(ctx context.Context) ([]*types.Message, error) {
	return message.Internal.ListMessage(ctx)
}

func (message *Message) UpdateMessageStateByCid(ctx context.Context, cid cid.Cid, state types.MessageState) (cid.Cid, error) {
	return message.Internal.UpdateMessageStateByCid(ctx, cid, state)
}

func (message *Message) UpdateMessageStateByID(ctx context.Context, id string, state types.MessageState) (string, error) {
	return message.Internal.UpdateMessageStateByID(ctx, id, state)
}

func (message *Message) UpdateAllFilledMessage(ctx context.Context) (int, error) {
	return message.Internal.UpdateAllFilledMessage(ctx)
}

func (message *Message) UpdateFilledMessageByID(ctx context.Context, id string) (string, error) {
	return message.Internal.UpdateFilledMessageByID(ctx, id)
}

func (message *Message) ReplaceMessage(ctx context.Context, id string, auto bool, maxFee string, gasLimit int64, gasPremium string, gasFeecap string) (cid.Cid, error) {
	return message.Internal.ReplaceMessage(ctx, id, auto, maxFee, gasLimit, gasPremium, gasFeecap)
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

func (message *Message) HasWallet(ctx context.Context, name string) (bool, error) {
	return message.Internal.HasWallet(ctx, name)
}

func (message *Message) ListRemoteWalletAddress(ctx context.Context, uuid types.UUID) ([]address.Address, error) {
	return message.Internal.ListRemoteWalletAddress(ctx, uuid)
}

func (message *Message) ListWallet(ctx context.Context) ([]*types.Wallet, error) {
	return message.Internal.ListWallet(ctx)
}

func (message *Message) DeleteWallet(ctx context.Context, name string) (string, error) {
	return message.Internal.DeleteWallet(ctx, name)
}

func (message *Message) UpdateWallet(ctx context.Context, wallet *types.Wallet) (string, error) {
	return message.Internal.UpdateWallet(ctx, wallet)
}

func (message *Message) SaveAddress(ctx context.Context, address *types.Address) (string, error) {
	return message.Internal.SaveAddress(ctx, address)
}

func (message *Message) GetAddress(ctx context.Context, addr address.Address) (*types.Address, error) {
	return message.Internal.GetAddress(ctx, addr)
}

func (message *Message) HasAddress(ctx context.Context, addr address.Address) (bool, error) {
	return message.Internal.HasAddress(ctx, addr)
}

func (message *Message) ListAddress(ctx context.Context) ([]*types.Address, error) {
	return message.Internal.ListAddress(ctx)
}

func (message *Message) UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) (address.Address, error) {
	return message.Internal.UpdateNonce(ctx, addr, nonce)
}

func (message *Message) DeleteAddress(ctx context.Context, addr address.Address) (address.Address, error) {
	return message.Internal.DeleteAddress(ctx, addr)
}

func (message *Message) ForbiddenAddress(ctx context.Context, addr address.Address) (address.Address, error) {
	return message.Internal.ForbiddenAddress(ctx, addr)
}

func (message *Message) ActiveAddress(ctx context.Context, addr address.Address) (address.Address, error) {
	return message.Internal.ActiveAddress(ctx, addr)
}

func (message *Message) UpdateSelectMsgNum(ctx context.Context, addr address.Address, num int) (address.Address, error) {
	return message.Internal.UpdateSelectMsgNum(ctx, addr, num)
}

func (message *Message) WaitMessage(ctx context.Context, id string, confidence uint64) (*types.Message, error) {
	tm := time.NewTicker(time.Second * 30)
	defer tm.Stop()

	doneCh := make(chan struct{}, 1)
	doneCh <- struct{}{}

	for {
		select {
		case <-doneCh:
			msg, err := message.Internal.GetMessageByUid(ctx, id)
			if err != nil {
				return nil, err
			}

			if msg.State == types.OnChainMsg && msg.Confidence > int64(confidence) {
				return msg, nil
			}
			continue
		case <-tm.C:
			doneCh <- struct{}{}
		case <-ctx.Done():
			return nil, xerrors.New("exit by client ")
		}
	}
}
