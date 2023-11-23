package testhelper

import (
	rand2 "crypto/rand"
	"math/rand"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	shared "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
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

func NewShareSignedMessage() *shared.SignedMessage {
	return &shared.SignedMessage{
		Message:   NewUnsignedMessage(),
		Signature: crypto.Signature{Type: crypto.SigTypeSecp256k1, Data: []byte(uuid.New().String())},
	}
}

func NewShareSignedMessages(count int) []*shared.SignedMessage {
	msgs := make([]*shared.SignedMessage, 0, count)
	for i := 0; i < count; i++ {
		msg := NewUnsignedMessage()
		msg.Nonce = uint64(i)
		msgs = append(msgs, &shared.SignedMessage{
			Message:   msg,
			Signature: crypto.Signature{Type: crypto.SigTypeSecp256k1, Data: []byte(uuid.New().String())},
		})
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
			GasOverEstimation: DefGasOverEstimation,
			GasOverPremium:    DefGasOverPremium,
			MaxFee:            big.NewInt(0),
		},
		Receipt:   &shared.MessageReceipt{ExitCode: -1},
		State:     types.UnFillMsg,
		CreatedAt: time.Now(),
	}
}

func NewUnsignedMessage() shared.Message {
	uid, _ := uuid.NewUUID()
	from, _ := address.NewActorAddress(uid[:])
	uid, _ = uuid.NewUUID()
	to, _ := address.NewActorAddress(uid[:])
	return shared.Message{
		From:       from,
		To:         to,
		Value:      big.NewInt(0),
		GasLimit:   0,
		GasFeeCap:  big.NewInt(0),
		GasPremium: big.NewInt(0),
	}
}

func RandNode() *types.Node {
	uuid := shared.NewUUID()
	uuidStr := uuid.String()
	return &types.Node{
		ID:    uuid,
		Name:  uuidStr,
		URL:   uuidStr,
		Token: uuidStr,
		Type:  types.NodeType(rand.Intn(3)),
	}
}

func RandNodes(count int) []*types.Node {
	nodes := make([]*types.Node, 0, count)
	for i := 0; i < count; i++ {
		nodes = append(nodes, RandNode())
	}

	return nodes
}

func MockSendSpecs() []*types.SendSpec {
	return []*types.SendSpec{
		nil,
		{
			GasOverEstimation: 1.25,
			MaxFee:            big.Mul(big.NewInt(DefGasUsed*100), DefGasFeeCap),
			GasOverPremium:    4.0,
		},
		{
			GasOverEstimation: 0,
			MaxFee:            big.NewInt(0),
			GasOverPremium:    0,
		},
		{
			GasOverEstimation: 0,
			GasOverPremium:    0,
		},
	}
}

func MockReplaceMessageParams() []*types.ReplacMessageParams {
	return []*types.ReplacMessageParams{
		{
			Auto: true,
		},
		{
			Auto:           true,
			GasOverPremium: 3.0,
			MaxFee:         big.Mul(big.NewInt(DefGasUsed*10), DefGasFeeCap),
		},
		{
			Auto:           true,
			GasOverPremium: 3.0,
			MaxFee:         big.Mul(big.NewInt(DefGasUsed/10), DefGasFeeCap),
		},
		{
			Auto:       false,
			GasLimit:   DefGasUsed * 10,
			GasPremium: big.Mul(DefGasFeeCap, big.NewInt(2)),
			GasFeecap:  big.Mul(DefGasPremium, big.NewInt(2)),
		},
	}
}

func GenBlockHead(miner address.Address, height abi.ChainEpoch, parents []cid.Cid) (*shared.BlockHeader, error) {
	data := make([]byte, 32)
	_, err := rand2.Read(data[:])
	if err != nil {
		return nil, err
	}
	c, err := abi.CidBuilder.Sum(data)
	if err != nil {
		return nil, err
	}
	return &shared.BlockHeader{
		Miner: miner,
		Ticket: &shared.Ticket{
			VRFProof: []byte("mock"),
		},
		ElectionProof:         nil,
		BeaconEntries:         nil,
		WinPoStProof:          nil,
		Parents:               parents,
		ParentWeight:          big.NewInt(100),
		Height:                height,
		ParentStateRoot:       c,
		ParentMessageReceipts: c,
		Messages:              c,
		BLSAggregate:          nil,
		Timestamp:             uint64(time.Now().Unix()),
		BlockSig:              nil,
		ForkSignaling:         0,
		ParentBaseFee:         DefBaseFee,
	}, nil
}

func GenTipset(height abi.ChainEpoch, width int, parents []cid.Cid) (*shared.TipSet, error) {
	var headers []*shared.BlockHeader
	for i := 0; i < width; i++ {
		addr, err := address.NewIDAddress(uint64(rand.Uint32()))
		if err != nil {
			return nil, err
		}
		blk, err := GenBlockHead(addr, height, parents)
		if err != nil {
			return nil, err
		}
		headers = append(headers, blk)
	}
	return shared.NewTipSet(headers)
}
