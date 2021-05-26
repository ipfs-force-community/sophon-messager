package mysql

import (
	"reflect"

	"github.com/filecoin-project/venus-messager/types"
)

var TMysqlMessage = reflect.TypeOf(&mysqlMessage{})
var TMessage = reflect.TypeOf(&types.Message{})

var TSharedParams = reflect.TypeOf(&types.SharedParams{})
var TMysqlSharedParams = reflect.TypeOf(&mysqlSharedParams{})

var TAddress = reflect.TypeOf(&types.Address{})
var TMysqlAddress = reflect.TypeOf(&mysqlAddress{})

var TNode = reflect.TypeOf(&types.Node{})
var TMysqlNode = reflect.TypeOf(&mysqlNode{})
