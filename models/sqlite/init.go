package sqlite

import (
	"fmt"
	"reflect"

	"github.com/filecoin-project/venus-messager/types"
)

var TMessage = reflect.TypeOf(&types.Message{})
var TSqliteMessage = reflect.TypeOf(&sqliteMessage{})

var TWallet = reflect.TypeOf(&types.Wallet{})
var TSqliteWallet = reflect.TypeOf(&sqliteWallet{})

var TAddress = reflect.TypeOf(&types.Address{})
var TSqliteAddress = reflect.TypeOf(&sqliteAddress{})

var TSqliteSharedParams = reflect.TypeOf(&sqliteSharedParams{})
var TSharedParams = reflect.TypeOf(&types.SharedParams{})

var TNode = reflect.TypeOf(&types.Node{})
var TSqliteNode = reflect.TypeOf(&sqliteNode{})

var TWalletAddress = reflect.TypeOf(&types.WalletAddress{})
var TSqliteWalletAddress = reflect.TypeOf(&sqliteWalletAddress{})

var TFeeConfig = reflect.TypeOf(&types.FeeConfig{})
var TSqliteFeeConfig = reflect.TypeOf(&sqliteFeeConfig{})

var ERRUnspportedMappingType = fmt.Errorf("unsupported mapping type")
