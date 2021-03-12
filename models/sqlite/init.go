package sqlite

import (
	"fmt"
	"reflect"

	"github.com/ipfs-force-community/venus-messager/types"
)

var TMessage = reflect.TypeOf(&types.Message{})
var TSqliteMessage = reflect.TypeOf(&sqliteMessage{})

var TWallet = reflect.TypeOf(&types.Wallet{})
var TSqliteWallet = reflect.TypeOf(&sqliteWallet{})

var TAddress = reflect.TypeOf(&types.Address{})
var TSqliteAddress = reflect.TypeOf(&sqliteAddress{})

var ERRUnspportedMappingType = fmt.Errorf("unsupported mapping type")
