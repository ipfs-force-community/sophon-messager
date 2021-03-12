package service

import (
	"context"
	"github.com/filecoin-project/venus/pkg/crypto"
	"net/http"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
)

type WalletClient struct {
	WalletList func(context.Context) ([]address.Address, error)
	WalletHas  func(ctx context.Context, address address.Address) (bool, error)
	WalletSign func(context.Context, address.Address, []byte) (*crypto.Signature, error)
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
