package client

import (
	"context"

	"github.com/ipfs-force-community/venus-messager/types"
)

type IMessager interface {
	PushMessage(ctx context.Context, msg *types.Message) (string, error)
	GetMessage(ctx context.Context, uuid string) (*types.Message, error)
	ListMessage(ctx context.Context) ([]*types.Message, error)

	SaveWallet(ctx context.Context, wallet *types.Wallet) (string, error)
	GetWallet(ctx context.Context, uuid string) (*types.Wallet, error)
	ListWallet(ctx context.Context) ([]*types.Wallet, error)

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

		SaveWallet func(ctx context.Context, wallet *types.Wallet) (string, error)
		GetWallet  func(ctx context.Context, uuid string) (*types.Wallet, error)
		ListWallet func(ctx context.Context) ([]*types.Wallet, error)

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
