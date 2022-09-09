package gateway

import (
	"context"
	"fmt"
	"sync"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/crypto"

	gatewayAPI "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	"github.com/filecoin-project/venus/venus-shared/types"
	gtypes "github.com/filecoin-project/venus/venus-shared/types/gateway"

	"github.com/filecoin-project/venus-messager/testhelper"
)

type MockWalletProxy struct {
	signers map[address.Address]struct{}

	l sync.Mutex
}

func NewMockWalletProxy() *MockWalletProxy {
	return &MockWalletProxy{
		signers: make(map[address.Address]struct{}),
	}
}

func (m *MockWalletProxy) AddAddress(addrs []address.Address) error {
	m.l.Lock()
	defer m.l.Unlock()

	for _, addr := range addrs {
		signerAddr := addr
		if addr.Protocol() == address.ID {
			newAddr, err := testhelper.ResolveIDAddr(addr)
			if err != nil {
				return err
			}
			signerAddr = newAddr
		}
		if _, ok := m.signers[signerAddr]; !ok {
			m.signers[signerAddr] = struct{}{}
		}
	}

	return nil
}

func (m *MockWalletProxy) RemoveAddress(ctx context.Context, addr address.Address) error {
	m.l.Lock()
	defer m.l.Unlock()

	if _, ok := m.signers[addr]; ok {
		delete(m.signers, addr)
	}

	return nil
}

func (m *MockWalletProxy) WalletHas(ctx context.Context, addr address.Address) (bool, error) {
	m.l.Lock()
	defer m.l.Unlock()

	_, ok := m.signers[addr]
	return ok, nil
}

func (m *MockWalletProxy) WalletSign(ctx context.Context, addr address.Address, toSign []byte, meta types.MsgMeta) (*crypto.Signature, error) {
	has, err := m.WalletHas(ctx, addr)
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

var _ gatewayAPI.IWalletClient = (*MockWalletProxy)(nil)
