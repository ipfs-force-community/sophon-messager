package gateway

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/venus-wallet/core"
	"github.com/ipfs-force-community/venus-common-utils/apiinfo"
	"github.com/ipfs-force-community/venus-gateway/walletevent"

	"github.com/filecoin-project/venus-messager/config"
)

type IWalletClient interface {
	WalletHas(ctx context.Context, supportAccount string, addr address.Address) (bool, error)
	WalletSign(ctx context.Context, account string, addr address.Address, toSign []byte, meta core.MsgMeta) (*crypto.Signature, error)

	ListWalletInfo(ctx context.Context) ([]*walletevent.WalletDetail, error)
}

// *api.MessageImp and *gateway.WalletClient both implement IWalletClient, so injection will fail
type IWalletCli struct {
	IWalletClient
}

type WalletClient struct {
	Internal struct {
		WalletHas  func(ctx context.Context, supportAccount string, addr address.Address) (bool, error)
		WalletSign func(ctx context.Context, account string, addr address.Address, toSign []byte, meta core.MsgMeta) (*crypto.Signature, error)

		ListWalletInfo func(ctx context.Context) ([]*walletevent.WalletDetail, error)
	}
}

func (w *WalletClient) WalletHas(ctx context.Context, supportAccount string, addr address.Address) (bool, error) {
	return w.Internal.WalletHas(ctx, supportAccount, addr)
}

func (w *WalletClient) WalletSign(ctx context.Context, account string, addr address.Address, toSign []byte, meta core.MsgMeta) (*crypto.Signature, error) {
	return w.Internal.WalletSign(ctx, account, addr, toSign, meta)
}

func (w *WalletClient) ListWalletInfo(ctx context.Context) ([]*walletevent.WalletDetail, error) {
	return w.Internal.ListWalletInfo(ctx)
}

func NewWalletClient(cfg *config.GatewayConfig) (IWalletClient, jsonrpc.ClientCloser, error) {
	return newWalletClient(context.Background(), cfg.Token, cfg.Url)
}

func newWalletClient(ctx context.Context, token, url string) (*WalletClient, jsonrpc.ClientCloser, error) {
	apiInfo := apiinfo.NewAPIInfo(url, token)
	addr, err := apiInfo.DialArgs("v0")
	if err != nil {
		return nil, nil, err
	}

	walletClient := WalletClient{}
	closer, err := jsonrpc.NewMergeClient(ctx, addr, "Gateway", []interface{}{&walletClient.Internal}, apiInfo.AuthHeader())

	return &walletClient, closer, err
}
