package types

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	venusTypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs/go-cid"
)

type MessageState int

const (
	UnKnown MessageState = iota
	UnFillMsg
	FillMsg
	OnChainMsg
	ExpireMsg
	ReplacedMsg
)

type MessageWithUID struct {
	UnsignedMessage venusTypes.UnsignedMessage
	ID              UUID
}

type Message struct {
	ID UUID

	UnsignedCid *cid.Cid
	SignedCid   *cid.Cid
	venusTypes.UnsignedMessage
	*crypto.Signature

	Height     int64
	Confidence int64
	Receipt    *venusTypes.MessageReceipt
	TipSetKey  venusTypes.TipSetKey

	Meta *MsgMeta

	State MessageState
}

func FromUnsignedMessage(unsignedMsg venusTypes.UnsignedMessage) *Message {
	return &Message{
		UnsignedMessage: unsignedMsg,
	}
}

type MsgMeta struct {
	ExpireEpoch       abi.ChainEpoch `json:"expireEpoch"`
	GasOverEstimation float64        `json:"gasOverEstimation"`
	MaxFee            big.Int        `json:"maxFee,omitempty"`
	MaxFeeCap         big.Int        `json:"maxFeeCap"`
}
