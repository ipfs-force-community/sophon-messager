package utils

import (
	"testing"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus/venus-shared/testutil"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMsgGroupByAddress(t *testing.T) {
	addrs := make([]address.Address, 10)
	msgCount := 200
	msgs := make([]*messager.Message, 0, msgCount)
	testutil.Provide(t, &addrs)

	addrMsgs := make(map[address.Address]map[string]*messager.Message)

	for i := 0; i < msgCount; i++ {
		addr := addrs[i%len(addrs)]
		id := uuid.New().String()
		msg := &messager.Message{ID: id, Message: types.Message{From: addr}}
		msgs = append(msgs, msg)
		if _, ok := addrMsgs[addr]; !ok {
			addrMsgs[addr] = make(map[string]*messager.Message)
		}
		addrMsgs[addr][id] = msg
	}

	for addr, msgList := range MsgGroupByAddress(msgs) {
		for _, msg := range msgList {
			assert.Equal(t, addrMsgs[addr][msg.ID], msg)
		}
		assert.Equal(t, len(addrMsgs[addr]), len(msgList))
	}
}
