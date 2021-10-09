package service

import (
	"sync"

	venusTypes "github.com/filecoin-project/venus/pkg/types"
)

func (tsCache *TipsetCache) RemoveTs(list []*tipsetFormat) {
	tsCache.l.Lock()
	defer tsCache.l.Unlock()
	for _, ts := range list {
		delete(tsCache.Cache, ts.Height)
	}
}

func (tsCache *TipsetCache) AddTs(list ...*venusTypes.TipSet) {
	tsCache.l.Lock()
	defer tsCache.l.Unlock()
	for _, ts := range list {
		tsCache.Cache[int64(ts.Height())] = ts
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
	if len(tsCache.Cache) < maxStoreTipsetCount {
		return
	}
	minHeight := tsCache.CurrHeight - maxStoreTipsetCount
	for _, v := range tsCache.Cache {
		if int64(v.Height()) < minHeight {
			delete(tsCache.Cache, int64(v.Height()))
		}
	}
}

func (tsCache *TipsetCache) ListTs() []*venusTypes.TipSet {
	tsCache.l.Lock()
	defer tsCache.l.Unlock()
	list := make([]*venusTypes.TipSet, 0, len(tsCache.Cache))
	for _, ts := range tsCache.Cache {
		list = append(list, ts)
	}

	return list
}

type idCidCache struct {
	cache map[string]string
	l     sync.Mutex
}

func (ic *idCidCache) Set(cid string, id string) {
	ic.l.Lock()
	defer ic.l.Unlock()
	ic.cache[cid] = id
}

func (ic *idCidCache) Get(cid string) (string, bool) {
	ic.l.Lock()
	defer ic.l.Unlock()
	id, ok := ic.cache[cid]

	return id, ok
}
