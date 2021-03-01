package mysql

import (
	"github.com/ipfs-force-community/venus-messager/types"
	"reflect"
)

var TMysqlMessage = reflect.TypeOf(&mysqlMessage{})
var TMessage = reflect.TypeOf(&types.Message{})

var TWallet = reflect.TypeOf(&types.Wallet{})
var TMysqlWallet = reflect.TypeOf(&mysqlWallet{})
