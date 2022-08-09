package gateway

import (
	"context"
	"fmt"
	"sync"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/venus/venus-shared/api/gateway/v1"
	"github.com/filecoin-project/venus/venus-shared/types"
	gtypes "github.com/filecoin-project/venus/venus-shared/types/gateway"

	"github.com/filecoin-project/venus-messager/testhelper"
)

type MockWalletProxy struct {
	accountAddrs map[string]map[address.Address]struct{}

	l sync.Mutex
}

func NewMockWalletProxy() *MockWalletProxy {
	return &MockWalletProxy{
		accountAddrs: make(map[string]map[address.Address]struct{}),
	}
}

func (m *MockWalletProxy) AddAddress(account string, addrs []address.Address) error {
	m.l.Lock()
	defer m.l.Unlock()

	currAddrs, ok := m.accountAddrs[account]
	if !ok {
		currAddrs = make(map[address.Address]struct{}, len(addrs))
	}
	for _, addr := range addrs {
		if addr.Protocol() == address.ID {
			newAddr, err := testhelper.ResolveIDAddr(addr)
			if err != nil {
				return err
			}
			currAddrs[newAddr] = struct{}{}
			continue
		}
		currAddrs[addr] = struct{}{}
	}
	m.accountAddrs[account] = currAddrs

	return nil
}

func (m *MockWalletProxy) RemoveAddress(account string, addrs []address.Address) error {
	m.l.Lock()
	defer m.l.Unlock()

	currAddrs, ok := m.accountAddrs[account]
	if ok {
		for _, addr := range addrs {
			if addr.Protocol() == address.ID {
				newAddr, err := testhelper.ResolveIDAddr(addr)
				if err != nil {
					return err
				}
				delete(currAddrs, newAddr)
				continue
			}
			delete(currAddrs, addr)
		}
	}

	return nil
}

func (m *MockWalletProxy) WalletHas(ctx context.Context, account string, addr address.Address) (bool, error) {
	m.l.Lock()
	defer m.l.Unlock()
	currAddrs, ok := m.accountAddrs[account]
	if !ok {
		return false, fmt.Errorf("not found account %v", account)
	}
	_, ok = currAddrs[addr]

	return ok, nil
}

func (m *MockWalletProxy) WalletSign(ctx context.Context, account string, addr address.Address, toSign []byte, meta types.MsgMeta) (*crypto.Signature, error) {
	has, err := m.WalletHas(ctx, account, addr)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("failed to found %s", addr)
	}
	return &crypto.Signature{
		Type: testhelper.AddressProtocolToSignType(addr.Protocol()),
		Data: append(toSign, addr.Bytes()...),
	}, nil
}

func (m *MockWalletProxy) ListWalletInfo(ctx context.Context) ([]*gtypes.WalletDetail, error) {
	panic("implement me")
}

func (m *MockWalletProxy) ListWalletInfoByWallet(ctx context.Context, wallet string) (*gtypes.WalletDetail, error) {
	panic("implement me")
}

var _ gateway.IWalletClient = (*MockWalletProxy)(nil)
