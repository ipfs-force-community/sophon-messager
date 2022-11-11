package utils

import (
	"testing"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus/venus-shared/testutil"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/stretchr/testify/assert"
)

func TestMsgsGroupByAddress(t *testing.T) {
	msgs := make([]*types.SignedMessage, 0, 10)
	testutil.Provide(t, &msgs, testutil.WithSliceLen(10))
	msgMap := make(map[address.Address][]*types.SignedMessage)
	for _, msg := range msgs {
		msgMap[msg.Message.From] = append(msgMap[msg.Message.From], msg)
	}

	for addr, msgs := range MsgsGroupByAddress(msgs) {
		tmp, ok := msgMap[addr]
		assert.True(t, ok)
		assert.Equal(t, tmp, msgs)
	}
}
