package types

import (
	"github.com/filecoin-project/venus/venus-shared/types/messager"
)

type State = messager.AddressState

const (
	Alive    = messager.AddressStateAlive
	Removing = messager.AddressStateRemoving
	Removed  = messager.AddressStateRemoved
	Forbiden = messager.AddressStateForbbiden
)

type Address = messager.Address

var StateToString = messager.AddressStateToString
