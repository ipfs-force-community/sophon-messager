package types

import (
	"github.com/filecoin-project/go-address"
	"sync"
)

type NonceMap struct {
	nonceMap map[address.Address]uint64
	lk       sync.RWMutex
}

func NewNonceMap() *NonceMap {
	return &NonceMap{
		nonceMap: make(map[address.Address]uint64),
		lk:       sync.RWMutex{},
	}
}

func (nonceMap *NonceMap) Get(addr address.Address) (uint64, bool) {
	nonceMap.lk.RLock()
	defer nonceMap.lk.RUnlock()
	if val, ok := nonceMap.nonceMap[addr]; ok {
		return val, ok
	}
	return 0, false
}

func (nonceMap *NonceMap) Add(addr address.Address, val uint64) {
	nonceMap.lk.RLock()
	defer nonceMap.lk.RUnlock()
	nonceMap.nonceMap[addr] = val
}
