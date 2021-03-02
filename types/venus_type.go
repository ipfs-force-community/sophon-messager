package types

import (
	"math"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/pkg/types"
)

type HeadChange struct {
	Type string
	Val  *types.TipSet
}

// SignedMessage contains a message and its signature
// TODO do not export these fields as it increases the chances of producing a
// `SignedMessage` with an empty signature.
type SignedMessage struct {
	Message   UnsignedMessage `json:"message"`
	Signature Signature       `json:"signature"`
}

type Signature struct {
	Type SigType
	Data []byte
}

type SigType byte

const (
	SigTypeUnknown = SigType(math.MaxUint8)

	SigTypeSecp256k1 = SigType(iota)
	SigTypeBLS
)

// UnsignedMessage is an exchange of information between two actors modeled
// as a function call.
type UnsignedMessage struct {
	Version uint64 `json:"version"`

	To   address.Address `json:"to"`
	From address.Address `json:"from"`
	// When receiving a message from a user account the nonce in
	// the message must match the expected nonce in the from actor.
	// This prevents replay attacks.
	Nonce uint64 `json:"nonce"`

	Value abi.TokenAmount `json:"value"`

	GasLimit   int64           `json:"gasLimit"`
	GasFeeCap  abi.TokenAmount `json:"gasFeeCap"`
	GasPremium abi.TokenAmount `json:"gasPremium"`

	Method abi.MethodNum `json:"method"`
	Params []byte        `json:"params"`
}
