package service

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/venus/pkg/chain"
	"reflect"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/venus-shared/actors/builtin"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-messager/types"
)

func (ms *MessageService) Send(ctx context.Context, params types.SendParams) (string, error) {
	var decParams []byte
	var err error

	if params.Method == builtin.MethodSend {
		return "", xerrors.Errorf("do not use it to send funds")
	}

	switch params.ParamsType {
	case types.ParamsJSON:
		decParams, err = ms.decodeTypedParamsFromJSON(ctx, params.To, params.Method, params.Params)
		if err != nil {
			return "", xerrors.Errorf("failed to decode json params: %w", err)
		}
	case types.ParamsHex:
		decParams, err = hex.DecodeString(params.Params)
		if err != nil {
			return "", xerrors.Errorf("failed to decode hex params: %w", err)
		}
	default:
		return "", xerrors.Errorf("unexpected param type %s", params.ParamsType)
	}

	uuid := types.NewUUID().String()
	msg := &types.Message{
		ID: uuid,
		Message: venusTypes.Message{
			From:  params.From,
			To:    params.To,
			Value: params.Val,

			Method: params.Method,
			Params: decParams,
		},
		State:      types.UnFillMsg,
		WalletName: params.Account,
		FromUser:   params.Account,
	}

	if params.GasPremium != nil {
		msg.Message.GasPremium = *params.GasPremium
	} else {
		msg.Message.GasPremium = abi.TokenAmount{Int: types.NewInt(0).Int}
	}
	if params.GasFeeCap != nil {
		msg.Message.GasFeeCap = *params.GasFeeCap
	} else {
		msg.Message.GasFeeCap = abi.TokenAmount{Int: types.NewInt(0).Int}
	}
	if params.GasLimit != nil {
		msg.Message.GasLimit = *params.GasLimit
	} else {
		msg.Message.GasLimit = 0
	}

	err = ms.pushMessage(ctx, msg)
	if err != nil {
		return "", err
	}

	return uuid, nil
}

func (ms *MessageService) decodeTypedParamsFromJSON(ctx context.Context, to address.Address, method abi.MethodNum, paramStr string) ([]byte, error) {
	act, err := ms.nodeClient.StateGetActor(ctx, to, venusTypes.EmptyTSK)
	if err != nil {
		return nil, err
	}

	methodMeta, found := chain.MethodsMap[act.Code][method]
	if !found {
		return nil, fmt.Errorf("method %d not found on actor %s", method, act.Code)
	}

	p := reflect.New(methodMeta.Params.Elem()).Interface().(cbg.CBORMarshaler)

	if err := json.Unmarshal([]byte(paramStr), p); err != nil {
		return nil, fmt.Errorf("unmarshaling input into params type: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := p.MarshalCBOR(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
