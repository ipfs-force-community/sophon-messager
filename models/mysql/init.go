package mysql

import (
	"reflect"

	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

var TMysqlMessage = reflect.TypeOf(&mysqlMessage{})
var TMessage = reflect.TypeOf(&types.Message{})

var TSharedParams = reflect.TypeOf(&types.SharedSpec{})
var TMysqlSharedParams = reflect.TypeOf(&mysqlSharedParams{})

var TAddress = reflect.TypeOf(&types.Address{})
var TMysqlAddress = reflect.TypeOf(&mysqlAddress{})

var TNode = reflect.TypeOf(&types.Node{})
var TMysqlNode = reflect.TypeOf(&mysqlNode{})
