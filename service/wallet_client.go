package service

import (
	"context"
	"github.com/filecoin-project/venus/pkg/crypto"
	"github.com/ipfs-force-community/venus-wallet/core"
	"net/http"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
)

type IWalletClient interface {
	WalletList(context.Context) ([]address.Address, error)
	WalletHas(context.Context, address.Address) (bool, error)
	WalletSign(ctx context.Context, signer address.Address, toSign []byte, meta core.MsgMeta) (*crypto.Signature, error)
}

var _ IWalletClient = (*WalletClient)(nil)

type WalletClient struct {
	Internal struct {
		WalletList func(context.Context) ([]address.Address, error)
		WalletHas  func(ctx context.Context, address address.Address) (bool, error)
		WalletSign func(ctx context.Context, signer address.Address, toSign []byte, meta core.MsgMeta) (*crypto.Signature, error)
	}
}

func (walletClient *WalletClient) WalletList(ctx context.Context) ([]address.Address, error) {
	return walletClient.Internal.WalletList(ctx)
}

func (walletClient *WalletClient) WalletHas(ctx context.Context, addr address.Address) (bool, error) {
	return walletClient.Internal.WalletHas(ctx, addr)
}

func (walletClient *WalletClient) WalletSign(ctx context.Context, signer address.Address, toSign []byte, meta core.MsgMeta) (*crypto.Signature, error) {
	return walletClient.Internal.WalletSign(ctx, signer, toSign, meta)
}

func newWalletClient(ctx context.Context, url, token string) (WalletClient, jsonrpc.ClientCloser, error) {
	headers := http.Header{}
	if len(token) != 0 {
		headers.Add("Authorization", "Bearer "+token)
	}
	addr, err := DialArgs(url)
	if err != nil {
		return WalletClient{}, nil, err
	}
	var res WalletClient
	closer, err := jsonrpc.NewMergeClient(ctx, addr, "Filecoin", []interface{}{&res.Internal}, headers)
	return res, closer, err
}
