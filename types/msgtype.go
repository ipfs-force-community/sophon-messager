package types

import (
	"github.com/filecoin-project/venus/venus-shared/types"
)

type MsgType = types.MsgType

const (
	MTUnknown = types.MTUnknown

	// Signing message CID. MsgMeta.Extra contains raw cbor message bytes
	MTChainMsg = types.MTChainMsg
)
