package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
)

func TestReadAndWriteTipset(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tsCache := &TipsetCache{Cache: map[int64]*venusTypes.TipSet{}, CurrHeight: 0}

	addTS := func(ts *venusTypes.TipSet) {
		tsCache.Cache[int64(ts.Height())] = ts
		tsCache.CurrHeight = int64(ts.Height())
	}

	msh := newMessageServiceHelper(ctx, t)
	count := 2
	currTS, err := msh.fullNode.ChainHead(ctx)
	assert.NoError(t, err)
	addTS(currTS)

	for i := 0; i < count; {
		ts, err := msh.fullNode.ChainHead(ctx)
		assert.NoError(t, err)
		if ts.Height() > currTS.Height() {
			addTS(currTS)
			currTS = ts
			i++
		}
		time.Sleep(msh.blockDelay)
	}

	filePath := filepath.Join(t.TempDir(), "tipset.json")
	assert.NoError(t, tsCache.Save(filePath))

	cache2 := newTipsetCache()
	err = cache2.Load(filePath)
	assert.NoError(t, err)
	assert.Equal(t, tsCache, cache2)
}
