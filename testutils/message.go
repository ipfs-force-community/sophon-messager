package testutils

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	types2 "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs-force-community/venus-messager/types"
)

func NewTestMsg() *types.Message {
	return &types.Message{
		Uid:             "44444",
		UnsignedMessage: NewTestUnsignedMsg(),
		// Signature:       &crypto.Signature{Type: crypto.SigTypeBLS, Data: []byte{1, 2, 3}},
		Meta: &types.MsgMeta{ExpireEpoch: 100,
			MaxFee: big.NewInt(10), GasOverEstimation: 0.5},
	}
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
