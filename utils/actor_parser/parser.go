package actor_parser

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/go-state-types/rt"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	exported0 "github.com/filecoin-project/specs-actors/actors/builtin/exported"
	exported2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/exported"
	exported3 "github.com/filecoin-project/specs-actors/v3/actors/builtin/exported"
	exported4 "github.com/filecoin-project/specs-actors/v4/actors/builtin/exported"
	exported5 "github.com/filecoin-project/specs-actors/v5/actors/builtin/exported"
	exported6 "github.com/filecoin-project/specs-actors/v6/actors/builtin/exported"
	exported7 "github.com/filecoin-project/specs-actors/v7/actors/builtin/exported"
	"github.com/filecoin-project/venus/venus-shared/types"
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
		return nil, xerrors.Errorf("registerActors actors v2 failed:%w", err)
	}
	if err = parser.registActors(exported3.BuiltinActors()...); err != nil {
		return nil, xerrors.Errorf("registerActors actors v3 failed:%w", err)
	}
	if err = parser.registActors(exported4.BuiltinActors()...); err != nil {
		return nil, xerrors.Errorf("registerActors actors v4 failed:%w", err)
	}
	if err = parser.registActors(exported5.BuiltinActors()...); err != nil {
		return nil, xerrors.Errorf("registerActors actors v5 failed:%w", err)
	}
	if err = parser.registActors(exported6.BuiltinActors()...); err != nil {
		return nil, xerrors.Errorf("registerActors actors v6 failed:%w", err)
	}
	if err = parser.registActors(exported7.BuiltinActors()...); err != nil {
		return nil, xerrors.Errorf("registerActors actors v7 failed:%w", err)
	}
	if err = parser.registActors(exported7.BuiltinActors()...); err != nil {
		return nil, xerrors.Errorf("registerActors actors v5 failed:%w", err)
	}
	return parser, nil
}

func (ms *MessagePaser) innerParseParams(ctx context.Context, method *Method, params []byte) (args interface{}, err error) {
	in := method.In()
	if err = in.UnmarshalCBOR(bytes.NewReader(params)); err != nil {
		return nil, xerrors.Errorf("unmarshalerCBOR msg params failed:%w", err)
	}
	return in, nil
}

func (ms *MessagePaser) innerParseReturn(ctx context.Context, method *Method, receipt *types.MessageReceipt) (ret interface{}, err error) {
	if receipt == nil {
		return "receipt is null", nil
	}

	if receipt.ExitCode != exitcode.Ok {
		return map[string]interface{}{"ExitCode": receipt.ExitCode,
			"Return": string(receipt.Return)}, nil
	}

	out := method.Out()
	if err = out.UnmarshalCBOR(bytes.NewReader(receipt.Return)); err != nil {
		return nil, xerrors.Errorf("unmarshalerCBOR msg returns failed:%w", err)
	}

	return out, nil
}

func (ms *MessagePaser) innerLookupMethod(ctx context.Context, to address.Address, num abi.MethodNum) (method *Method, err error) {
	if num == builtin.MethodSend {
		return nil, xerrors.Errorf("no method type for 'MethodSend'")
	}

	var actor *types.Actor
	if actor, err = ms.getter.StateGetActor(ctx, to, types.EmptyTSK); err != nil {
		return nil, xerrors.Errorf("get actor(%s) failed:%w", to.String(), err)
	}

	var actorType *Actor
	var find bool

	if actorType, find = ms.lookUpActor(actor.Code); !find {
		return nil, xerrors.Errorf("actor code(%s) not registed", actor.Code.String())
	}

	if method, find = actorType.lookUpMethod(int(num)); !find {
		return nil, xerrors.Errorf("actor:%s method(%d) not exist", actorType.Name, num)
	}
	return method, nil
}

func (ms *MessagePaser) LookUpMsgMethod(ctx context.Context, msg *types.Message) (method *Method, err error) {
	return ms.innerLookupMethod(ctx, msg.To, msg.Method)
}

func (ms *MessagePaser) ParseParams(ctx context.Context, msg *types.Message) (args interface{}, err error) {
	method, err := ms.LookUpMsgMethod(ctx, msg)
	if err != nil {
		return nil, xerrors.Errorf("lookup method failed:%w", err)
	}
	return ms.innerParseParams(ctx, method, msg.Params)
}

func (ms *MessagePaser) ParseReturn(ctx context.Context, msg *types.Message, receipt *types.MessageReceipt) (ret interface{}, err error) {
	method, err := ms.LookUpMsgMethod(ctx, msg)
	if err != nil {
		return nil, xerrors.Errorf("lookup method failed:%w", err)
	}
	return ms.innerParseReturn(ctx, method, receipt)
}

func (ms *MessagePaser) DecodeParamsFromJSON(ctx context.Context, to address.Address, num abi.MethodNum, jParam []byte) ([]byte, error) {
	method, err := ms.innerLookupMethod(ctx, to, num)
	if err != nil {
		return nil, err
	}

	in := method.In()
	if err := json.Unmarshal(jParam, in); err != nil {
		return nil, xerrors.Errorf("unmarshal jsoned param to : %s failed: %w", method.inType.String(), err)
	}

	buf := new(bytes.Buffer)
	if err := in.MarshalCBOR(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
