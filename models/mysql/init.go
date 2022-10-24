package mysql

import (
	"reflect"

	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

var (
	TMysqlMessage = reflect.TypeOf(&mysqlMessage{})
	TMessage      = reflect.TypeOf(&types.Message{})
)

var (
	TSharedParams      = reflect.TypeOf(&types.SharedSpec{})
	TMysqlSharedParams = reflect.TypeOf(&mysqlSharedParams{})
)

var (
	TAddress      = reflect.TypeOf(&types.Address{})
	TMysqlAddress = reflect.TypeOf(&mysqlAddress{})
)

var (
	TNode      = reflect.TypeOf(&types.Node{})
	TMysqlNode = reflect.TypeOf(&mysqlNode{})
)
