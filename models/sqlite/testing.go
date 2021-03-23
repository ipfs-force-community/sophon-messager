package sqlite

import (
	"encoding/json"
	"math/rand"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	types2 "github.com/filecoin-project/venus/pkg/types"
	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/google/uuid"

	"github.com/ipfs-force-community/venus-messager/types"
)

func NewSignedMessages(count int) []*types.Message {
	msgs := make([]*types.Message, 0, count)
	for i := 0; i < count; i++ {
		msg := NewMessage()
		msg.Nonce = uint64(i)
		msg.Signature = &crypto.Signature{Type: crypto.SigTypeSecp256k1, Data: []byte(uuid.New().String())}
		unsignedCid := msg.UnsignedMessage.Cid()
		msg.UnsignedCid = &unsignedCid
		signedCid := (&venustypes.SignedMessage{
			Message:   msg.UnsignedMessage,
			Signature: *msg.Signature,
		}).Cid()
		msg.SignedCid = &signedCid
		msgs = append(msgs, msg)
	}

	return msgs
}

func NewMessage() *types.Message {
	return &types.Message{
		ID:              types.NewUUID().String(),
		UnsignedMessage: NewUnsignedMessage(),
		Meta: &types.MsgMeta{
			ExpireEpoch:       100,
			MaxFee:            big.NewInt(10),
			GasOverEstimation: 0.5,
		},
		Receipt: &venustypes.MessageReceipt{ExitCode: -1},
	}
}

func NewUnsignedMessage() types2.UnsignedMessage {
	from, _ := address.NewFromString("f01234")
	to, _ := address.NewFromString("f01235")
	return types2.UnsignedMessage{
		From:       from,
		To:         to,
		Value:      big.NewInt(rand.Int63n(1024)),
		GasLimit:   rand.Int63n(100),
		GasFeeCap:  abi.NewTokenAmount(2000),
		GasPremium: abi.NewTokenAmount(1024),
	}
}

func ObjectToString(i interface{}) string {
	res, _ := json.MarshalIndent(i, "", " ")
	return string(res)
}
