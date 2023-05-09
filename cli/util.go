package cli

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/venus-messager/cli/tablewriter"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/filecoin-project/venus/venus-shared/utils"
	"github.com/ipfs/go-cid"
)

var msgTw = tablewriter.New(
	tablewriter.Col("ID"),
	tablewriter.Col("To"),
	tablewriter.Col("From"),
	tablewriter.Col("Nonce"),
	tablewriter.Col("Value"),
	tablewriter.Col("GasLimit"),
	tablewriter.Col("GasFeeCap"),
	tablewriter.Col("GasPremium"),
	tablewriter.Col("Method"),
	tablewriter.Col("State"),
	tablewriter.Col("ExitCode"),
	tablewriter.Col("CreateAt"),
)

func outputWithTable(msgs []*types.Message, verbose bool, nodeAPI v1.FullNode) error {
	for _, msgT := range msgs {
		msg := transformMessage(msgT, nodeAPI)
		val := venusTypes.MustParseFIL(msg.Msg.Value.String() + "attofil").String()
		row := map[string]interface{}{
			"ID":         msg.ID,
			"To":         msg.Msg.To,
			"From":       msg.Msg.From,
			"Nonce":      msg.Msg.Nonce,
			"Value":      val,
			"GasLimit":   msg.Msg.GasLimit,
			"GasFeeCap":  msg.Msg.GasFeeCap,
			"GasPremium": msg.Msg.GasPremium,
			"Method":     msg.Msg.Method,
			"State":      msg.State,
			"ErrorMsg":   msg.ErrorMsg,
			"CreateAt":   msg.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if !verbose {
			if from := msg.Msg.From.String(); len(from) > 9 {
				row["From"] = from[:9] + "..."
			}
			if to := msg.Msg.To.String(); len(to) > 9 {
				row["To"] = to[:9] + "..."
			}
			if len(msg.ID) > 36 {
				row["ID"] = msg.ID[:36] + "..."
			}
			if len(val) > 6 {
				row["Value"] = val[:6] + "..."
			}
		}
		if msg.Receipt != nil {
			row["ExitCode"] = msg.Receipt.ExitCode
		}
		msgTw.Write(row)
	}

	buf := new(bytes.Buffer)
	if err := msgTw.Flush(buf); err != nil {
		return err
	}
	fmt.Println(buf)
	return nil
}

type msgTmp struct {
	Version    uint64
	To         address.Address
	From       address.Address
	Nonce      uint64
	Value      abi.TokenAmount
	GasLimit   int64
	GasFeeCap  abi.TokenAmount
	GasPremium abi.TokenAmount
	Method     string
	Params     []byte
}

type receipt struct {
	ExitCode exitcode.ExitCode
	Return   string
	GasUsed  int64
}

type message struct {
	ID string

	UnsignedCid *cid.Cid
	SignedCid   *cid.Cid
	Msg         msgTmp
	Signature   *crypto.Signature

	Height     int64
	Confidence int64
	Receipt    *receipt
	TipSetKey  venusTypes.TipSetKey

	Meta *types.SendSpec

	WalletName string
	ErrorMsg   string
	State      string

	UpdatedAt time.Time
	CreatedAt time.Time
}

func transformMessage(msg *types.Message, nodeAPI v1.FullNode) *message {
	if msg == nil {
		return nil
	}

	m := &message{
		ID:          msg.ID,
		UnsignedCid: msg.UnsignedCid,
		SignedCid:   msg.SignedCid,
		Signature:   msg.Signature,
		Height:      msg.Height,
		Confidence:  msg.Confidence,
		TipSetKey:   msg.TipSetKey,
		Meta:        msg.Meta,
		WalletName:  msg.WalletName,
		State:       msg.State.String(),
		ErrorMsg:    msg.ErrorMsg,

		UpdatedAt: msg.UpdatedAt,
		CreatedAt: msg.CreatedAt,
	}
	if msg.Receipt != nil {
		m.Receipt = &receipt{
			ExitCode: msg.Receipt.ExitCode,
			Return:   string(msg.Receipt.Return),
			GasUsed:  msg.Receipt.GasUsed,
		}
	}
	m.Msg = msgTmp{
		Version:    msg.Version,
		To:         msg.To,
		From:       msg.From,
		Nonce:      msg.Nonce,
		Value:      msg.Value,
		GasLimit:   msg.GasLimit,
		GasFeeCap:  msg.GasFeeCap,
		GasPremium: msg.GasPremium,
		Method:     methodToStr(nodeAPI, msg.Message),
		Params:     msg.Params,
	}

	return m
}

func methodToStr(nodeAPI v1.FullNode, msg venusTypes.Message) string {
	methodStr, err := func() (string, error) {
		actor, err := nodeAPI.StateGetActor(context.Background(), msg.To, venusTypes.EmptyTSK)
		if err != nil {
			return "", fmt.Errorf("get actor(%s) failed: %v", msg.To, err)
		}

		methodMeta, found := utils.MethodsMap[actor.Code][msg.Method]
		if !found {
			return "", fmt.Errorf("actor(%v) method(%d) not exist", actor.Code, msg.Method)
		}

		return methodMeta.Name, nil
	}()
	if err != nil {
		fmt.Println("failed to parse message method to string: ", err)
		return msg.Method.String()
	}

	return methodStr
}

func resolveBuiltinMethodName(methodType types.MethodType) string {
	methodMeta, found := utils.MethodsMap[methodType.Code][methodType.Method]
	if !found {
		return ""
	}

	return methodMeta.Name
}
