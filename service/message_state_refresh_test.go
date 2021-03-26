package service

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadAndWriteTipset(t *testing.T) {
	tsCache := &TipsetCache{Cache: map[int64]*tipsetFormat{}, CurrHeight: 0}
	tsCache.Cache[0] = &tipsetFormat{
		Key:    "00000",
		Height: 0,
	}
	tsCache.Cache[3] = &tipsetFormat{
		Key:    "33333",
		Height: 3,
	}
	tsCache.Cache[2] = &tipsetFormat{
		Key:    "22222",
		Height: 2,
	}
	tsCache.CurrHeight = 3

	filePath := "./test_read_write_tipset.txt"
	defer func() {
		//assert.NoError(t, os.Remove(filePath))
	}()
	err := updateTipsetFile(filePath, tsCache)
	assert.NoError(t, err)

	result, err := readTipsetFile(filePath)
	assert.NoError(t, err)
	assert.Len(t, result.Cache, 3)

	var tsList tipsetList
	for _, c := range result.Cache {
		tsList = append(tsList, c)
	}
	t.Logf("before sort %+v", tsList)

	sort.Sort(tsList)
	t.Logf("after sort %+v", tsList)
	assert.Equal(t, tsList[1].Height, int64(2))
}
