package gateway

import (
	"context"
	"fmt"
	"sync"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/crypto"

	"github.com/filecoin-project/venus/venus-shared/api/gateway/v1"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	gtypes "github.com/filecoin-project/venus/venus-shared/types/gateway"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/log"
)

type WalletProxy struct {
	clients map[string]gateway.IWalletClient
	logger  *log.Logger
}

func (w *WalletProxy) fastSelectAvaWalletClient(ctx context.Context, addr address.Address) (gateway.IWalletClient, error) {
	var g = &sync.WaitGroup{}
	var ch = make(chan gateway.IWalletClient, 1)

	for url, c := range w.clients {
		g.Add(1)
		go func(url string, c gateway.IWalletClient) {
			has, err := c.WalletHas(ctx, addr)
			if err != nil {
				w.logger.Errorf("fastSelectAvaWalletClient, call %s:'WalletHas' failed:%s", url, err)
			}
			if has {
				ch <- c
			}

			g.Done()
		}(url, c)
	}

	go func() {
		g.Wait()
		close(ch)
	}()

	c, isok := <-ch
	if !isok || c == nil {
		return nil, fmt.Errorf("can't find a wallet, signer address: %s", addr)
	}

	return c, nil
}

func (w *WalletProxy) WalletHas(ctx context.Context, addr address.Address) (bool, error) {
	c, err := w.fastSelectAvaWalletClient(ctx, addr)
	return c != nil, err
}

func (w *WalletProxy) WalletSign(ctx context.Context, addr address.Address, toSign []byte, meta venusTypes.MsgMeta) (*crypto.Signature, error) {
	c, err := w.fastSelectAvaWalletClient(ctx, addr)
	if err != nil {
		return nil, err
	}

	return c.WalletSign(ctx, addr, toSign, meta)
}

func (w *WalletProxy) ListWalletInfo(context.Context) ([]*gtypes.WalletDetail, error) {
	panic("implement me")
}

func (w *WalletProxy) ListWalletInfoByWallet(context.Context, string) (*gtypes.WalletDetail, error) {
	panic("implement me")
}

func NewWalletClient(ctx context.Context,
	cfg *config.GatewayConfig,
	logger *log.Logger,
) (*WalletProxy, jsonrpc.ClientCloser, error) {
	var proxy = &WalletProxy{
		clients: make(map[string]gateway.IWalletClient),
		logger:  logger,
	}
	var ctx = context.Background()

	var closers []jsonrpc.ClientCloser

	for _, url := range cfg.Url {
		c, cls, err := gateway.DialIGatewayRPC(ctx, url, cfg.Token, nil)

		if err != nil {
			return nil, nil, fmt.Errorf("create geteway client with url:%s failed: %w", url, err)
		}

		proxy.clients[url] = c
		closers = append(closers, cls)
	}

	if len(proxy.clients) == 0 {
		return nil, nil, fmt.Errorf("can't create any gateway client, please check 'GatewayConfig'")
	}

	totalCloser := func() {
		for _, closer := range closers {
			closer()
		}
	}

	return proxy, totalCloser, nil
}

var _ gateway.IWalletClient = &WalletProxy{}
