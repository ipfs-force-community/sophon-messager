package service

import (
	"context"
	"sync"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"golang.org/x/xerrors"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/types"
)

// 轮询去扫本地wallet表，根据wallet表中URL和token生成 Client，再根据Client轮询去获取地址列表

type Addresses struct {
	addressService *AddressService
	walletService  *WalletService
	nodeClient     *NodeClient
	log            *logrus.Logger

	cfg          *config.AddressConfig
	addrNonceMap map[string]uint64

	l sync.Mutex
}

func ListenAddressChange(lc fx.Lifecycle, addressService *AddressService, logger *logrus.Logger, nodeClient *NodeClient, cfg *config.AddressConfig) *Addresses {
	a := &Addresses{
		addressService: addressService,
		log:            logger,
		nodeClient:     nodeClient,
		cfg:            cfg,
		addrNonceMap:   make(map[string]uint64),
		l:              sync.Mutex{},
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := a.getLocalAddressAndNonce(); err != nil {
				return xerrors.Errorf("get local address and nonce failed: %v", err)
			}
			a.listenAddressChange(ctx)

			return nil
		},
	})

	return a
}

func (a Addresses) getLocalAddressAndNonce() error {
	addrs, err := a.addressService.ListAddress(context.Background())
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		a.setNonce(addr.Addr, addr.Nonce)
	}

	return nil
}

func (a Addresses) listenAddressChange(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(a.cfg.RemoteWalletSweepInterval * time.Second)
		for {
			select {
			case <-ticker.C:
				for key, cli := range a.walletService.walletClients {
					addrs, err := cli.WalletList(ctx)
					if err != nil {
						a.log.Errorf("get wallet list failed, url: %s, err: %v", key, err)
						continue
					}
					a.ProcessAddress(addrs)
				}
			default:
			}
		}
	}()
}

func (a Addresses) ProcessAddress(addrs []address.Address) {
	for _, addr := range addrs {
		if _, ok := a.addrNonceMap[addr.String()]; ok {
			continue
		}

		var nonce uint64
		actor, err := a.nodeClient.GetActor(context.Background(), addr)
		if err != nil {
			a.log.Warnf("get actor failed, addr: %s, err: %v", addr, err)
		} else {
			nonce = actor.Nonce
		}

		ta := &types.Address{
			Addr:      addr.String(),
			Nonce:     nonce,
			UpdatedAt: time.Now(),
		}

		if err := a.SetNonceToLocal(addr.String(), actor.Nonce); err != nil {
			a.log.Errorf("set nonce failed addr: %v, err: %v", ta, err)
		}
	}
}

func (a Addresses) GetNonce(addr string) uint64 {
	a.l.Lock()
	defer a.l.Unlock()
	return a.addrNonceMap[addr]
}

func (a Addresses) setNonce(addr string, nonce uint64) {
	a.l.Lock()
	defer a.l.Unlock()
	a.addrNonceMap[addr] = nonce
}

func (a Addresses) SetNonceToLocal(addr string, nonce uint64) error {
	a.setNonce(addr, nonce)

	_, err := a.addressService.SaveAddress(context.Background(), &types.Address{
		Addr:      addr,
		Nonce:     nonce,
		UpdatedAt: time.Now(),
	})

	return err
}
