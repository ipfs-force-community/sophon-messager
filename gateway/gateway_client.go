package gateway

import (
	"context"
	"fmt"
	"sync"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/crypto"

	gatewayAPI "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	gtypes "github.com/filecoin-project/venus/venus-shared/types/gateway"
	logging "github.com/ipfs/go-log/v2"

	"github.com/filecoin-project/venus-messager/config"
)

var log = logging.Logger("wallet-proxy")

type cacheKey string

func newCacheKey(addr address.Address) cacheKey {
	return cacheKey("walletClientCache:" + addr.String())
}

type WalletProxy struct {
	clients map[string]gatewayAPI.IWalletClient

	mutx                sync.RWMutex
	avaliabeClientCache map[cacheKey]gatewayAPI.IWalletClient
}

func (w *WalletProxy) putCache(addr address.Address, client gatewayAPI.IWalletClient) {
	w.mutx.Lock()
	defer w.mutx.Unlock()
	w.avaliabeClientCache[newCacheKey(addr)] = client
}

func (w *WalletProxy) delCache(addr address.Address) bool {
	key := newCacheKey(addr)
	w.mutx.Lock()
	defer w.mutx.Unlock()
	_, exist := w.avaliabeClientCache[key]
	if exist {
		delete(w.avaliabeClientCache, key)
	}
	return exist
}

func (w *WalletProxy) getCachedClient(addr address.Address) gatewayAPI.IWalletClient {
	w.mutx.RLock()
	defer w.mutx.RUnlock()

	key := newCacheKey(addr)
	return w.avaliabeClientCache[key]
}

// todo: think about 'fastSelectAvaGatewayClient' was called parallelly
func (w *WalletProxy) fastSelectAvaGatewayClient(ctx context.Context, addr address.Address, accounts []string) (gatewayAPI.IWalletClient, error) {
	var g = &sync.WaitGroup{}
	var ch = make(chan gatewayAPI.IWalletClient, 1)
	for url, c := range w.clients {
		g.Add(1)
		go func(url string, c gatewayAPI.IWalletClient) {
			has, err := c.WalletHas(ctx, addr, accounts)
			if err != nil {
				log.Errorf("fastSelectAvaClient, call %s:'WalletHas' failed:%s", url, err)
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
		return nil, fmt.Errorf("can't find a wallet, address: %s", addr.String())
	}

	w.putCache(addr, c)
	return c, nil
}

func (w *WalletProxy) WalletHas(ctx context.Context, addr address.Address, accounts []string) (bool, error) {
	c := w.getCachedClient(addr)
	if c != nil {
		return true, nil
	}
	c, err := w.fastSelectAvaGatewayClient(ctx, addr, accounts)
	return c != nil, err
}

func (w *WalletProxy) WalletSign(ctx context.Context, addr address.Address, accounts []string, toSign []byte, meta venusTypes.MsgMeta) (*crypto.Signature, error) {
	var err error
	var useCachedClient bool

	c := w.getCachedClient(addr)

	if c == nil {
		if c, err = w.fastSelectAvaGatewayClient(ctx, addr, accounts); err != nil {
			return nil, err
		}
	} else {
		useCachedClient = true
	}

	var s *crypto.Signature
	if s, err = c.WalletSign(ctx, addr, accounts, toSign, meta); err != nil {
		if useCachedClient {
			log.Warnf("sign with cached client failed:%s, will re-SelectAvaliableClient, and retry",
				err.Error())

			w.delCache(addr)

			if c, err = w.fastSelectAvaGatewayClient(ctx, addr, accounts); err != nil {
				return nil, err
			}
			s, err = c.WalletSign(ctx, addr, accounts, toSign, meta)
		}
	}

	return s, err
}

func (w *WalletProxy) ListWalletInfo(context.Context) ([]*gtypes.WalletDetail, error) {
	panic("implement me")
}

func (w *WalletProxy) ListWalletInfoByWallet(context.Context, string) (*gtypes.WalletDetail, error) {
	panic("implement me")
}

func NewWalletClient(ctx context.Context,
	cfg *config.GatewayConfig,
) (*WalletProxy, jsonrpc.ClientCloser, error) {
	var proxy = &WalletProxy{
		clients:             make(map[string]gatewayAPI.IWalletClient),
		avaliabeClientCache: make(map[cacheKey]gatewayAPI.IWalletClient),
	}

	var closers []jsonrpc.ClientCloser
	for _, url := range cfg.Url {
		c, cls, err := gatewayAPI.DialIGatewayRPC(ctx, url, cfg.Token, nil)

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

var _ gatewayAPI.IWalletClient = &WalletProxy{}
