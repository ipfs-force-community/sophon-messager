package gateway

import (
	"context"
	"fmt"
	"sync"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/crypto"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/venus-common-utils/apiinfo"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/log"
)

type IWalletClient interface {
	WalletHas(ctx context.Context, supportAccount string, addr address.Address) (bool, error)
	WalletSign(ctx context.Context, account string, addr address.Address, toSign []byte, meta venusTypes.MsgMeta) (*crypto.Signature, error)
}

// IWalletCli *api.MessageImp and *gateway.WalletClient both implement IWalletClient, so injection will fail
type IWalletCli struct {
	IWalletClient
}

type WalletClient struct {
	Internal struct {
		WalletHas  func(ctx context.Context, supportAccount string, addr address.Address) (bool, error)
		WalletSign func(ctx context.Context, account string, addr address.Address, toSign []byte, meta venusTypes.MsgMeta) (*crypto.Signature, error)
	}
}

type WalletProxy struct {
	clients map[string]*WalletClient
	logger  *log.Logger

	mutx                sync.RWMutex
	avaliabeClientCache map[cacheKey]*WalletClient
}

type cacheKey string

func newCacheKey(account string, addr address.Address) cacheKey {
	return cacheKey("walletClientCache:" + account + addr.String())
}

func (w *WalletProxy) putCache(account string, addr address.Address, client *WalletClient) {
	w.mutx.Lock()
	defer w.mutx.Unlock()
	w.avaliabeClientCache[newCacheKey(account, addr)] = client
}

func (w *WalletProxy) delCache(account string, addr address.Address) bool {
	key := newCacheKey(account, addr)
	w.mutx.Lock()
	defer w.mutx.Unlock()
	_, exist := w.avaliabeClientCache[key]
	if exist {
		delete(w.avaliabeClientCache, key)
	}
	return exist
}

func (w *WalletProxy) getCachedClient(account string, addr address.Address) *WalletClient {
	key := newCacheKey(account, addr)
	w.mutx.RLock()
	defer w.mutx.RUnlock()
	return w.avaliabeClientCache[key]
}

// todo: think about 'fastSelectAvaClient' was called parallelly,
//  input the same params('account', 'address')
func (w *WalletProxy) fastSelectAvaClient(ctx context.Context, account string, addr address.Address) (*WalletClient, error) {
	var g = &sync.WaitGroup{}
	var ch = make(chan *WalletClient, 1)
	for url, c := range w.clients {
		g.Add(1)
		go func(url string, c *WalletClient) {
			has, err := c.WalletHas(ctx, account, addr)
			if err != nil {
				w.logger.Errorf("fastSelectAvaClient, call %s:'WalletHas' failed:%s", url, err)
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
		return nil, fmt.Errorf("can't find a wallet, account: %s address: %s", account, addr.String())
	}

	w.putCache(account, addr, c)
	return c, nil
}

func (w *WalletProxy) WalletHas(ctx context.Context, supportAccount string, addr address.Address) (bool, error) {
	c := w.getCachedClient(supportAccount, addr)
	if c != nil {
		return true, nil
	}
	c, err := w.fastSelectAvaClient(ctx, supportAccount, addr)
	return c != nil, err
}

func (w *WalletProxy) WalletSign(ctx context.Context, account string,
	addr address.Address, toSign []byte, meta venusTypes.MsgMeta) (*crypto.Signature, error) {
	var err error
	var useCachedClient bool

	c := w.getCachedClient(account, addr)

	if c == nil {
		if c, err = w.fastSelectAvaClient(ctx, account, addr); err != nil {
			return nil, err
		}
	} else {
		useCachedClient = true
	}

	var s *crypto.Signature
	if s, err = c.WalletSign(ctx, account, addr, toSign, meta); err != nil {
		if useCachedClient {

			w.logger.Warnf("sign with cached client failed:%s, will re-SelectAvaliableClient, and retry",
				err.Error())

			w.delCache(account, addr)

			if c, err = w.fastSelectAvaClient(ctx, account, addr); err != nil {
				return nil, err
			}

			s, err = c.WalletSign(ctx, account, addr, toSign, meta)
		}
	}

	return s, err
}

func (w *WalletClient) WalletHas(ctx context.Context, supportAccount string, addr address.Address) (bool, error) {
	return w.Internal.WalletHas(ctx, supportAccount, addr)
}

func (w *WalletClient) WalletSign(ctx context.Context, account string, addr address.Address, toSign []byte, meta venusTypes.MsgMeta) (*crypto.Signature, error) {
	return w.Internal.WalletSign(ctx, account, addr, toSign, meta)
}

func NewWalletClient(cfg *config.GatewayConfig, logger *log.Logger) (*WalletProxy, jsonrpc.ClientCloser, error) {
	var proxy = &WalletProxy{
		clients:             make(map[string]*WalletClient),
		avaliabeClientCache: make(map[cacheKey]*WalletClient),
		logger:              logger}
	var ctx = context.Background()

	var closers []jsonrpc.ClientCloser

	for _, url := range cfg.Url {
		c, cls, err := newWalletClient(ctx, cfg.Token, url)

		if err != nil {
			return nil, nil, fmt.Errorf("create geteway client with url:%s failed:%w", url, err)
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
