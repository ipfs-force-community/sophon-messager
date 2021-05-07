package service

import (
	"context"
	"sync"

	ffi "github.com/filecoin-project/filecoin-ffi"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-wallet/core"
	"github.com/filecoin-project/venus/pkg/crypto"
	"golang.org/x/xerrors"
)

type MemWallet struct {
	lk   sync.Mutex
	keys map[address.Address]ffi.PrivateKey
}

//just for test
func NewMemWallet() *MemWallet {
	wallet := &MemWallet{
		lk:   sync.Mutex{},
		keys: make(map[address.Address]ffi.PrivateKey),
	}

	priv := [32]byte{49, 245, 245, 84, 117, 222, 231, 108, 225, 166, 56, 151, 45, 39, 212, 139, 56, 185, 70, 67, 33, 240, 229, 164, 166, 79, 23, 26, 109, 48, 109, 84}
	pub := ffi.PrivateKeyPublicKey(priv)

	addr, _ := address.NewBLSAddress(pub[:])
	wallet.keys[addr] = priv
	return wallet
}

func (memWallet *MemWallet) WalletList(ctx context.Context) ([]address.Address, error) {
	memWallet.lk.Lock()
	defer memWallet.lk.Unlock()
	var addrs []address.Address
	for addr := range memWallet.keys {
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

func (memWallet *MemWallet) WalletHas(ctx context.Context, addr address.Address) (bool, error) {
	memWallet.lk.Lock()
	defer memWallet.lk.Unlock()
	_, ok := memWallet.keys[addr]
	return ok, nil
}

func (memWallet *MemWallet) WalletSign(ctx context.Context, addr address.Address, data []byte, meta core.MsgMeta) (*crypto.Signature, error) {
	memWallet.lk.Lock()
	defer memWallet.lk.Unlock()
	priv, ok := memWallet.keys[addr]
	if !ok {
		return nil, xerrors.Errorf("no privkey of address %s", addr)
	}
	sig := ffi.PrivateKeySign(priv, data)
	return &crypto.Signature{
		Type: crypto.SigTypeBLS,
		Data: sig[:],
	}, nil
}
