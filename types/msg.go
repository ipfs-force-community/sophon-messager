package types

import (
	"github.com/filecoin-project/venus/venus-shared/types/messager"
)

type MessageState = messager.MessageState

const (
	UnKnown     = messager.UnKnown
	UnFillMsg   = messager.UnFillMsg
	FillMsg     = messager.FillMsg
	OnChainMsg  = messager.OnChainMsg
	FailedMsg   = messager.FailedMsg
	ReplacedMsg = messager.ReplacedMsg
	NoWalletMsg = messager.NoWalletMsg
)

type MessageWithUID = messager.MessageWithUID

type Message = messager.Message

var FromUnsignedMessage = messager.FromUnsignedMessage

type MsgMeta = messager.SendSpec

var MsgStateToString = messager.MessageStateToString
