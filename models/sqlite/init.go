package sqlite

import (
	"reflect"

	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

var (
	TMessage       = reflect.TypeOf(&types.Message{})
	TSqliteMessage = reflect.TypeOf(&sqliteMessage{})
)

var (
	TAddress       = reflect.TypeOf(&types.Address{})
	TSqliteAddress = reflect.TypeOf(&sqliteAddress{})
)

var (
	TSqliteSharedParams = reflect.TypeOf(&sqliteSharedParams{})
	TSharedParams       = reflect.TypeOf(&types.SharedSpec{})
)

var (
	TNode       = reflect.TypeOf(&types.Node{})
	TSqliteNode = reflect.TypeOf(&sqliteNode{})
)
