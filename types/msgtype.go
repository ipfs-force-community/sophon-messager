package types

type MsgType string

const (
	MTUnknown = MsgType("unknown")

	// Signing message CID. MsgMeta.Extra contains raw cbor message bytes
	MTChainMsg = MsgType("message")
)
