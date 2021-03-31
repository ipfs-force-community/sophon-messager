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
	NoWalletMsg
)

//						---> ExpireMsg <------
//					    |					 |
// 				UnFillMsg ---------------> FillMsg --------> OnChainMsg
//						|					 |
//		 NoWalletMsg <---				     ---->ReplacedMsg
//

type MessageWithUID struct {
	UnsignedMessage venusTypes.UnsignedMessage
	ID              string
}

type Message struct {
	ID string

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

func MsgStateToString(state MessageState) string {
	switch state {
	case UnFillMsg:
		return "UnFillMsg"
	case FillMsg:
		return "FillMsg"
	case OnChainMsg:
		return "OnChainMsg"
	case ExpireMsg:
		return "ExpireMsg"
	case ReplacedMsg:
		return "ReplacedMsg"
	case NoWalletMsg:
		return "NoWalletMsg"
	default:
		return "UnKnown"
	}
}
