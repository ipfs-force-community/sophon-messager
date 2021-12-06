package actor_parser

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/cbor"
	"github.com/ipfs/go-cid"
	"reflect"
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

func (m *Method) In() cbor.Er {
	return reflect.New(m.inType).Interface().(cbor.Er)
}

func (m *Method) Out() cbor.Er {
	return reflect.New(m.outType).Interface().(cbor.Er)
}
