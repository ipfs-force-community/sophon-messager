package sqlite

import (
	"github.com/ipfs-force-community/venus-messager/types"
	"reflect"
)

var TMessage = reflect.TypeOf(&types.Message{})
var TSqliteMessage = reflect.TypeOf(&sqliteMessage{})
