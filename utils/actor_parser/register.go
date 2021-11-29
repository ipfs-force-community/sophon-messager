package actor_parser

import (
	"reflect"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
)

type Method struct {
	Name    string
	Num     int
	inType  reflect.Type
	outType reflect.Type
}

type Actor struct {
	Code    cid.Cid
	Name    string
	methods map[abi.MethodNum]*Method
}

func (actor *Actor) lookUpMethod(num int) (*Method, bool) {
	method, exist := actor.methods[abi.MethodNum(num)]
	return method, exist
}
