package gateway

import (
	"context"
	"net/http"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/venus-wallet/core"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/utils"
)

type IWalletClient interface {
	WalletHas(ctx context.Context, supportAccount string, addr address.Address) (bool, error)
	WalletSign(ctx context.Context, account string, addr address.Address, toSign []byte, meta core.MsgMeta) (*crypto.Signature, error)
}

// *api.MessageImp and *gateway.WalletClient both implement IWalletClient, so injection will fail
type IWalletCli struct {
	IWalletClient
}

type WalletClient struct {
	Internal struct {
		WalletHas  func(ctx context.Context, supportAccount string, addr address.Address) (bool, error)
		WalletSign func(ctx context.Context, account string, addr address.Address, toSign []byte, meta core.MsgMeta) (*crypto.Signature, error)
	}
}

func (w *WalletClient) WalletHas(ctx context.Context, supportAccount string, addr address.Address) (bool, error) {
	return w.Internal.WalletHas(ctx, supportAccount, addr)
}

func (w *WalletClient) WalletSign(ctx context.Context, account string, addr address.Address, toSign []byte, meta core.MsgMeta) (*crypto.Signature, error) {
	return w.Internal.WalletSign(ctx, account, addr, toSign, meta)
}

func NewWalletClient(cfg *config.GatewayConfig) (IWalletClient, jsonrpc.ClientCloser, error) {
	return newWalletClient(context.Background(), cfg.Token, cfg.Url)
}

func newWalletClient(ctx context.Context, token, url string) (*WalletClient, jsonrpc.ClientCloser, error) {
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+token)

	walletClient := WalletClient{}

	addr, err := utils.DialArgs(url)
	if err != nil {
		return nil, nil, err
	}
	closer, err := jsonrpc.NewMergeClient(ctx, addr, "Filecoin", []interface{}{&walletClient.Internal}, headers)

	return &walletClient, closer, err
}
