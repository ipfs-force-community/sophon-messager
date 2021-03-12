package utils

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	types2 "github.com/filecoin-project/venus/pkg/types"
	"github.com/google/uuid"

	"github.com/ipfs-force-community/venus-messager/types"
)

func NewTestMsg() *types.Message {
	return &types.Message{
		ID:              types.NewUUID(),
		UnsignedMessage: NewTestUnsignedMsg(),
		// Signature:       &crypto.Signature{Type: crypto.SigTypeBLS, Data: []byte{1, 2, 3}},
		Meta: &types.MsgMeta{ExpireEpoch: 100,
			MaxFee: big.NewInt(10), GasOverEstimation: 0.5},
	}
}

func NewTestSignedMsgs(count int) []*types.Message {
	msgs := make([]*types.Message, 0, count)
	for i := 0; i < count; i++ {
		msg := NewTestMsg()
		msg.Signature = &crypto.Signature{Type: crypto.SigTypeSecp256k1, Data: []byte(uuid.New().String())}
		msgs = append(msgs, msg)
	}

	return msgs
}

func NewTestUnsignedMsg() types2.UnsignedMessage {
	from, _ := address.NewFromString("f01234")
	to, _ := address.NewFromString("f01235")
	return types2.UnsignedMessage{
		From:       from,
		To:         to,
		Value:      big.NewInt(1024),
		GasLimit:   100,
		GasFeeCap:  abi.NewTokenAmount(2000),
		GasPremium: abi.NewTokenAmount(1024),
	}
}
