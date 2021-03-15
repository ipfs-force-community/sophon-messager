package client

import (
	"context"
	"github.com/filecoin-project/go-address"
	"golang.org/x/xerrors"
	"time"

	"github.com/ipfs-force-community/venus-messager/types"
)

type IMessager interface {
	WaitMessage(ctx context.Context, uuid string) (*types.Message, error)
	PushMessage(ctx context.Context, msg *types.Message) (string, error)
	GetMessage(ctx context.Context, uuid string) (*types.Message, error)
	ListMessage(ctx context.Context) ([]*types.Message, error)

	SaveWallet(ctx context.Context, wallet *types.Wallet) (string, error)
	GetWallet(ctx context.Context, uuid string) (*types.Wallet, error)
	ListWallet(ctx context.Context) ([]*types.Wallet, error)
	ListWalletAddress(ctx context.Context, name string) ([]address.Address, error)

	SaveAddress(ctx context.Context, address *types.Address) (string, error)
	GetAddress(ctx context.Context, addr string) (*types.Address, error)
	ListAddress(ctx context.Context) ([]*types.Address, error)
}

var _ IMessager = (*Message)(nil)

type Message struct {
	Internal struct {
		PushMessage func(ctx context.Context, msg *types.Message) (string, error)
		GetMessage  func(ctx context.Context, uuid string) (*types.Message, error)
		ListMessage func(ctx context.Context) ([]*types.Message, error)

		SaveWallet        func(ctx context.Context, wallet *types.Wallet) (string, error)
		GetWallet         func(ctx context.Context, uuid string) (*types.Wallet, error)
		ListWallet        func(ctx context.Context) ([]*types.Wallet, error)
		ListWalletAddress func(ctx context.Context, name string) ([]address.Address, error)

		SaveAddress func(ctx context.Context, address *types.Address) (string, error)
		GetAddress  func(ctx context.Context, addr string) (*types.Address, error)
		ListAddress func(ctx context.Context) ([]*types.Address, error)
	}
}

func (message *Message) PushMessage(ctx context.Context, msg *types.Message) (string, error) {
	return message.Internal.PushMessage(ctx, msg)
}

func (message *Message) GetMessage(ctx context.Context, uuid string) (*types.Message, error) {
	return message.Internal.GetMessage(ctx, uuid)
}

func (message *Message) ListMessage(ctx context.Context) ([]*types.Message, error) {
	return message.Internal.ListMessage(ctx)
}

func (message *Message) SaveWallet(ctx context.Context, wallet *types.Wallet) (string, error) {
	return message.Internal.SaveWallet(ctx, wallet)
}

func (message *Message) GetWallet(ctx context.Context, uuid string) (*types.Wallet, error) {
	return message.Internal.GetWallet(ctx, uuid)
}

func (message *Message) ListWalletAddress(ctx context.Context, name string) ([]address.Address, error) {
	return message.Internal.ListWalletAddress(ctx, name)
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

func (message *Message) ListAddress(ctx context.Context) ([]*types.Address, error) {
	return message.Internal.ListAddress(ctx)
}

func (message *Message) WaitMessage(ctx context.Context, uuid string) (*types.Message, error) {
	tm := time.NewTicker(time.Second * 30)
	defer tm.Stop()

	for {
		select {
		case <-tm.C:
			msg, err := message.Internal.GetMessage(ctx, uuid)
			if err != nil {
				return nil, err
			}
			if msg.State == types.OnChainMsg {
				return msg, nil
			}
			continue
		case <-ctx.Done():
			return nil, xerrors.New("exit by client ")
		}
	}
}
