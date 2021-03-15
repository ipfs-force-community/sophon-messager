package service

import (
	"context"
	"github.com/filecoin-project/venus/pkg/crypto"
	"net/http"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
)

type IWalletClient interface {
	WalletList(context.Context) ([]address.Address, error)
	WalletHas(context.Context, address.Address) (bool, error)
	WalletSign(context.Context, address.Address, []byte) (*crypto.Signature, error)
}

var _ IWalletClient = (*WalletClient)(nil)

type WalletClient struct {
	Internal struct {
		WalletList func(context.Context) ([]address.Address, error)
		WalletHas  func(ctx context.Context, address address.Address) (bool, error)
		WalletSign func(context.Context, address.Address, []byte) (*crypto.Signature, error)
	}
}

func (walletClient *WalletClient) WalletList(ctx context.Context) ([]address.Address, error) {
	return walletClient.Internal.WalletList(ctx)
}

func (walletClient *WalletClient) WalletHas(ctx context.Context, addr address.Address) (bool, error) {
	return walletClient.Internal.WalletHas(ctx, addr)
}

func (walletClient *WalletClient) WalletSign(ctx context.Context, addr address.Address, data []byte) (*crypto.Signature, error) {
	return walletClient.Internal.WalletSign(ctx, addr, data)
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
	closer, err := jsonrpc.NewMergeClient(ctx, addr, "Filecoin", []interface{}{&res}, headers)
	return res, closer, err
}
