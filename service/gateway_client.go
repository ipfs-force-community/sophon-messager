package service

import (
	"context"
	"net/http"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/venus-wallet/core"

	"github.com/filecoin-project/venus-messager/config"
)

type WalletEventClient struct {
	WalletHas  func(ctx context.Context, supportAccount string, addr address.Address) (bool, error)
	WalletSign func(ctx context.Context, account string, addr address.Address, toSign []byte, meta core.MsgMeta) (*crypto.Signature, error)
}

type GatewayClient struct {
	WalletClient *WalletEventClient
	Close        jsonrpc.ClientCloser
}

func NewGatewayClient(cfg *config.GatewayConfig) (*GatewayClient, error) {
	cli, close, err := newWalletEventClient(context.Background(), cfg.Token, cfg.Url)
	if err != nil {
		return nil, err
	}

	return &GatewayClient{
		WalletClient: cli,
		Close:        close,
	}, nil
}

func newWalletEventClient(ctx context.Context, token, url string) (*WalletEventClient, jsonrpc.ClientCloser, error) {
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+token)
	walletEventCli := &WalletEventClient{}
	addr, err := DialArgs(url)
	if err != nil {
		return nil, nil, err
	}
	closer, err := jsonrpc.NewMergeClient(ctx, addr, "Filecoin", []interface{}{walletEventCli}, headers)

	return walletEventCli, closer, err
}
