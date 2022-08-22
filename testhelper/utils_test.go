package testhelper

import (
	"testing"

	"github.com/filecoin-project/venus/venus-shared/testutil"
	"github.com/stretchr/testify/assert"
)

func TestResolveIDAddr(t *testing.T) {
	loop := 100
	for i := 0; i < loop; i++ {
		addr := testutil.IDAddressProvider()(t)
		newAddr, err := ResolveIDAddr(addr)
		assert.NoError(t, err)

		newAddr2, err := ResolveIDAddr(addr)
		assert.NoError(t, err)

		assert.Equal(t, newAddr, newAddr2)
	}
}
