package service

import (
	"encoding/json"
	"os"
	"sync"

	venusTypes "github.com/filecoin-project/venus/venus-shared/types"

	"github.com/filecoin-project/venus-messager/utils"
)

const maxStoreTipsetCount = 900

type TipsetCache struct {
	Cache       map[int64]*venusTypes.TipSet
	CurrHeight  int64
	NetworkName string

	l sync.Mutex
}

func newTipsetCache() *TipsetCache {
	return &TipsetCache{
		Cache: make(map[int64]*venusTypes.TipSet, maxStoreTipsetCount),
	}
}

func (tsCache *TipsetCache) Load(path string) error {
	b, err := utils.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var tmp TipsetCache
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	tsCache.Cache = tmp.Cache
	tsCache.CurrHeight = tmp.CurrHeight
	tsCache.NetworkName = tmp.NetworkName

	return nil
}

func (tsCache *TipsetCache) Add(list ...*venusTypes.TipSet) {
	tsCache.l.Lock()
	defer tsCache.l.Unlock()
	for _, ts := range list {
		tsCache.Cache[int64(ts.Height())] = ts
	}
}

func (tsCache *TipsetCache) reduce() {
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

func (tsCache *TipsetCache) List() []*venusTypes.TipSet {
	tsCache.l.Lock()
	defer tsCache.l.Unlock()
	list := make([]*venusTypes.TipSet, 0, len(tsCache.Cache))
	for _, ts := range tsCache.Cache {
		list = append(list, ts)
	}

	return list
}

// original data will be cleared
func (tsCache *TipsetCache) Save(filePath string) error {
	tsCache.reduce()
	return utils.WriteFile(filePath, tsCache)
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
