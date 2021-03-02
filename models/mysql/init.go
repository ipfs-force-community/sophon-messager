package mysql

import (
	"reflect"

	"github.com/ipfs-force-community/venus-messager/types"
)

var TMysqlMessage = reflect.TypeOf(&mysqlMessage{})
var TMessage = reflect.TypeOf(&types.Message{})

var TWallet = reflect.TypeOf(&types.Wallet{})
var TMysqlWallet = reflect.TypeOf(&mysqlWallet{})

var TAddress = reflect.TypeOf(&types.Address{})
var TMysqlAddress = reflect.TypeOf(&mysqlAddress{})
