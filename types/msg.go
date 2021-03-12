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
	UnKnown = iota
	UnFillMsg
	FillMsg
	OnChainMsg
	ExpireMsg
)

type Message struct {
	ID UUID

	UnsignedCid *cid.Cid
	SignedCid   *cid.Cid
	venusTypes.UnsignedMessage
	*crypto.Signature

	Height  uint64
	Receipt *venusTypes.MessageReceipt

	Meta *MsgMeta

	State MessageState // 消息状态
}

type MsgMeta struct {
	ExpireEpoch       abi.ChainEpoch `json:"expireEpoch"`
	GasOverEstimation float64        `json:"gasOverEstimation"`
	MaxFee            big.Int        `json:"maxFee,omitempty"`
	MaxFeeCap         big.Int        `json:"maxFeeCap"`
}
