package types

import (
	"time"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs/go-cid"
)

type MessageState int

const (
	UnKnown MessageState = iota
	UnFillMsg
	FillMsg
	OnChainMsg
	FailedMsg
	ReplacedMsg
	NoWalletMsg
)

//						---> FailedMsg <------
//					    |					 |
// 				UnFillMsg ---------------> FillMsg --------> OnChainMsg
//						|					 |
//		 NoWalletMsg <---				     ---->ReplacedMsg
//

type MessageWithUID struct {
	UnsignedMessage venusTypes.Message
	ID              string
}

type Message struct {
	ID string

	UnsignedCid *cid.Cid
	SignedCid   *cid.Cid
	venusTypes.Message
	Signature *crypto.Signature

	Height     int64
	Confidence int64
	Receipt    *venusTypes.MessageReceipt
	TipSetKey  venusTypes.TipSetKey
	Meta       *MsgMeta
	WalletName string
	FromUser   string

	State MessageState

	CreatedAt time.Time
	UpdatedAt time.Time
}

func FromUnsignedMessage(unsignedMsg venusTypes.Message) *Message {
	return &Message{
		Message: unsignedMsg,
	}
}

type MsgMeta struct {
	ExpireEpoch       abi.ChainEpoch `json:"expireEpoch"`
	GasOverEstimation float64        `json:"gasOverEstimation"`
	MaxFee            big.Int        `json:"maxFee,omitempty"`
	MaxFeeCap         big.Int        `json:"maxFeeCap"`
}

func MsgStateToString(state MessageState) string {
	switch state {
	case UnFillMsg:
		return "UnFillMsg"
	case FillMsg:
		return "FillMsg"
	case OnChainMsg:
		return "OnChainMsg"
	case FailedMsg:
		return "Failed"
	case ReplacedMsg:
		return "ReplacedMsg"
	case NoWalletMsg:
		return "NoWalletMsg"
	default:
		return "UnKnown"
	}
}
