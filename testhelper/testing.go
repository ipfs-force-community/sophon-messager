package testhelper

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	shared "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/google/uuid"
)

func NewSignedMessages(count int) []*types.Message {
	msgs := make([]*types.Message, 0, count)
	for i := 0; i < count; i++ {
		msg := NewMessage()
		msg.Nonce = uint64(i)
		msg.Signature = &crypto.Signature{Type: crypto.SigTypeSecp256k1, Data: []byte(uuid.New().String())}
		unsignedCid := msg.Message.Cid()
		msg.UnsignedCid = &unsignedCid
		signedCid := (&shared.SignedMessage{
			Message:   msg.Message,
			Signature: *msg.Signature,
		}).Cid()
		msg.SignedCid = &signedCid
		msgs = append(msgs, msg)
	}

	return msgs
}

func NewMessages(count int) []*types.Message {
	msgs := make([]*types.Message, count)
	for i := 0; i < count; i++ {
		msgs[i] = NewMessage()
	}

	return msgs
}

func NewMessage() *types.Message {
	return &types.Message{
		ID:      shared.NewUUID().String(),
		Message: NewUnsignedMessage(),
		Meta: &types.SendSpec{
			ExpireEpoch:       100,
			MaxFee:            big.NewInt(10),
			GasOverEstimation: 0.5,
		},
		Receipt: &shared.MessageReceipt{ExitCode: -1},
		State:   types.UnFillMsg,
	}
}

func NewUnsignedMessage() shared.Message {
	rand.Seed(time.Now().Unix())
	uid, _ := uuid.NewUUID()
	from, _ := address.NewActorAddress(uid[:])
	uid, _ = uuid.NewUUID()
	to, _ := address.NewActorAddress(uid[:])
	return shared.Message{
		From:       from,
		To:         to,
		Value:      big.NewInt(rand.Int63n(1024)),
		GasLimit:   rand.Int63n(100),
		GasFeeCap:  abi.NewTokenAmount(2000),
		GasPremium: abi.NewTokenAmount(1024),
	}
}

func ObjectToString(obj interface{}) string {
	res, err := json.Marshal(obj)
	if err != nil {
		panic(fmt.Errorf("marshal failed %v", err))
	}
	return string(res)
}
