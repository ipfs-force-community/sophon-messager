package sqlite

import (
	"reflect"

	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

var TMessage = reflect.TypeOf(&types.Message{})
var TSqliteMessage = reflect.TypeOf(&sqliteMessage{})

var TAddress = reflect.TypeOf(&types.Address{})
var TSqliteAddress = reflect.TypeOf(&sqliteAddress{})

var TSqliteSharedParams = reflect.TypeOf(&sqliteSharedParams{})
var TSharedParams = reflect.TypeOf(&types.SharedSpec{})

var TNode = reflect.TypeOf(&types.Node{})
var TSqliteNode = reflect.TypeOf(&sqliteNode{})
