package service

import (
	"sync"

	"github.com/ipfs-force-community/venus-messager/types"
)

func (tsCache *TipsetCache) RemoveTs(list []*tipsetFormat) {
	tsCache.l.Lock()
	defer tsCache.l.Unlock()
	for _, ts := range list {
		delete(tsCache.Cache, ts.Height)
	}
}

func (tsCache *TipsetCache) AddTs(list ...*tipsetFormat) {
	tsCache.l.Lock()
	defer tsCache.l.Unlock()
	for _, ts := range list {
		tsCache.Cache[ts.Height] = ts
	}
}

func (tsCache *TipsetCache) ExistTs(height int64) bool {
	tsCache.l.Lock()
	defer tsCache.l.Unlock()
	_, ok := tsCache.Cache[height]

	return ok
}

func (tsCache *TipsetCache) ReduceTs() {
	tsCache.l.Lock()
	defer tsCache.l.Unlock()
	minHeight := tsCache.CurrHeight - maxStoreTipsetCount
	for _, v := range tsCache.Cache {
		if v.Height < minHeight {
			delete(tsCache.Cache, v.Height)
		}
	}
}

func (tsCache *TipsetCache) ListTs() tipsetList {
	tsCache.l.Lock()
	defer tsCache.l.Unlock()
	var list tipsetList
	for _, ts := range tsCache.Cache {
		list = append(list, ts)
	}

	return list
}

type idCidCache struct {
	cache map[string]types.UUID
	l     sync.Mutex
}

func (ic *idCidCache) Set(cid string, id types.UUID) {
	ic.l.Lock()
	defer ic.l.Unlock()
	ic.cache[cid] = id
}

func (ic *idCidCache) Get(cid string) (types.UUID, bool) {
	ic.l.Lock()
	defer ic.l.Unlock()
	id, ok := ic.cache[cid]

	return id, ok
}
