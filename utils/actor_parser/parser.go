package actor_parser

import (
	"bytes"
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/cbor"
	"github.com/filecoin-project/go-state-types/rt"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	exported0 "github.com/filecoin-project/specs-actors/actors/builtin/exported"
	exported2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/exported"
	exported3 "github.com/filecoin-project/specs-actors/v3/actors/builtin/exported"
	exported4 "github.com/filecoin-project/specs-actors/v4/actors/builtin/exported"
	exported5 "github.com/filecoin-project/specs-actors/v5/actors/builtin/exported"
	exported6 "github.com/filecoin-project/specs-actors/v6/actors/builtin/exported"
	"github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs/go-cid"
	"golang.org/x/xerrors"
	"reflect"
	"strings"
)

type ActorGetter interface {
	StateGetActor(context.Context, address.Address, types.TipSetKey) (*types.Actor, error)
	StateLookupID(context.Context, address.Address, types.TipSetKey) (address.Address, error)
}

type MessagePaser struct {
	getter ActorGetter
	actors map[cid.Cid]*Actor
}

func (parser *MessagePaser) registActors(actors ...rt.VMActor) error {
	for _, actor := range actors {
		if err := parser.registActor(actor); err != nil {
			return err
		}
	}
	return nil
}

func (parser *MessagePaser) registActor(actor rt.VMActor) error {
	if parser.actors == nil {
		parser.actors = make(map[cid.Cid]*Actor)
	}

	funcs := actor.Exports()

	pkgPath := strings.Split(reflect.TypeOf(actor).PkgPath(), "/")

	var actorType = Actor{
		Name:    pkgPath[len(pkgPath)-1],
		Code:    actor.Code(),
		methods: make(map[abi.MethodNum]*Method),
	}

	indirect := func(p reflect.Type) reflect.Type {
		for p.Kind() == reflect.Ptr {
			p = p.Elem()
		}
		return p
	}

	for idx, f := range funcs {
		if f == nil {
			continue
		}
		mt := reflect.TypeOf(f)
		var in, out reflect.Type
		iNum := mt.NumIn()
		oNum := mt.NumOut()

		if iNum > 0 {
			in = indirect(mt.In(iNum - 1))
		}

		if oNum > 0 {
			out = indirect(mt.Out(0))
		}

		actorType.methods[abi.MethodNum(idx)] = &Method{
			Name:    mt.Name(),
			Num:     idx,
			inType:  in,
			outType: out,
		}
	}

	parser.actors[actorType.Code] = &actorType

	return nil
}

func (parser *MessagePaser) lookUpActor(code cid.Cid) (*Actor, bool) {
	actor, exist := parser.actors[code]
	return actor, exist
}

func NewMessageParser(getter ActorGetter) (*MessagePaser, error) {
	parser := &MessagePaser{getter: getter}
	var err error
	if err = parser.registActors(exported0.BuiltinActors()...); err != nil {
		return nil, xerrors.Errorf("registerActors actors v0 failed:%w", err)
	}
	if err = parser.registActors(exported2.BuiltinActors()...); err != nil {
		return nil, xerrors.Errorf("registerActors actors v1 failed:%w", err)
	}
	if err = parser.registActors(exported3.BuiltinActors()...); err != nil {
		return nil, xerrors.Errorf("registerActors actors v2 failed:%w", err)
	}
	if err = parser.registActors(exported4.BuiltinActors()...); err != nil {
		return nil, xerrors.Errorf("registerActors actors v3 failed:%w", err)
	}
	if err = parser.registActors(exported5.BuiltinActors()...); err != nil {
		return nil, xerrors.Errorf("registerActors actors v4 failed:%w", err)
	}
	if err = parser.registActors(exported6.BuiltinActors()...); err != nil {
		return nil, xerrors.Errorf("registerActors actors v5 failed:%w", err)
	}
	return parser, nil
}

func (ms *MessagePaser) ParseMessage(ctx context.Context, msg *types.Message, receipt *types.MessageReceipt) (args interface{}, ret interface{}, err error) {
	if int(msg.Method) == int(builtin.MethodSend) {
		return nil, nil, nil
	}
	var actor *types.Actor
	if actor, err = ms.getter.StateGetActor(ctx, msg.To, types.EmptyTSK); err != nil {
		return nil, nil, xerrors.Errorf("get actor(%s) failed:%w", msg.To.String(), err)
	}

	var actorType, method, find = (*Actor)(nil), (*Method)(nil), false

	if actorType, find = ms.lookUpActor(actor.Code); !find {
		return nil, nil, xerrors.Errorf("actor code(%s) not registed", actor.Code.String())
	}

	if method, find = actorType.lookUpMethod(int(msg.Method)); !find {
		return nil, nil, xerrors.Errorf("actor:%s method(%d) not exist", actorType.Name, msg.Method)
	}

	in := reflect.New(method.inType).Interface()

	if unmarshaler, isok := in.(cbor.Unmarshaler); isok {
		if err = unmarshaler.UnmarshalCBOR(bytes.NewReader(msg.Params)); err != nil {
			return nil, nil, xerrors.Errorf("unmarshalerCBOR msg params failed:%w", err)
		}
	}

	var out interface{}
	if receipt != nil {
		out = reflect.New(method.outType).Interface()
		if unmarshaler, isok := out.(cbor.Unmarshaler); isok {
			if err = unmarshaler.UnmarshalCBOR(bytes.NewReader(receipt.ReturnValue)); err != nil {
				return nil, nil, xerrors.Errorf("unmarshalerCBOR msg returns failed:%w", err)
			}
		}
	}

	return in, out, nil
}
