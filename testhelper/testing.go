package testhelper

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	shared "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func NewSignedMessages(count int) []*types.Message {
	msgs := make([]*types.Message, 0, count)
	for i := 0; i < count; i++ {
		msg := NewMessage()
		msg.Nonce = uint64(i)
		msg.Signature = &crypto.Signature{Type: crypto.SigTypeSecp256k1, Data: []byte(uuid.New().String())}
		unsignedCid := msg.Message.Cid()
		msg.UnsignedCid = &unsignedCid
		signedCid := (&shared.SignedMessage{
			Message:   msg.Message,
			Signature: *msg.Signature,
		}).Cid()
		msg.SignedCid = &signedCid
		msgs = append(msgs, msg)
	}

	return msgs
}

func NewMessages(count int) []*types.Message {
	msgs := make([]*types.Message, count)
	for i := 0; i < count; i++ {
		msgs[i] = NewMessage()
	}

	return msgs
}

func NewMessage() *types.Message {
	return &types.Message{
		ID:      shared.NewUUID().String(),
		Message: NewUnsignedMessage(),
		Meta: &types.SendSpec{
			ExpireEpoch:       100,
			MaxFee:            big.NewInt(10),
			GasOverEstimation: 0.5,
		},
		Receipt:   &shared.MessageReceipt{ExitCode: -1},
		State:     types.UnFillMsg,
		CreatedAt: time.Now(),
	}
}

func NewUnsignedMessage() shared.Message {
	rand.Seed(time.Now().Unix())
	uid, _ := uuid.NewUUID()
	from, _ := address.NewActorAddress(uid[:])
	uid, _ = uuid.NewUUID()
	to, _ := address.NewActorAddress(uid[:])
	return shared.Message{
		From:       from,
		To:         to,
		Value:      big.NewInt(rand.Int63n(1024)),
		GasLimit:   rand.Int63n(100),
		GasFeeCap:  abi.NewTokenAmount(2000),
		GasPremium: abi.NewTokenAmount(1024),
	}
}

func ObjectToString(obj interface{}) string {
	res, err := json.Marshal(obj)
	if err != nil {
		panic(fmt.Errorf("marshal failed %v", err))
	}
	return string(res)
}

const timeFormat = "2006-01-02 15:04:05.999"

func Equal(t *testing.T, expect, actual interface{}) {
	expectRV := reflect.ValueOf(expect)
	expectRT := reflect.TypeOf(expect)
	actualRV := reflect.ValueOf(actual)
	assert.Equal(t, expectRV.Kind(), actualRV.Kind())
	if expectRV.Kind() == reflect.Ptr {
		expectRV = expectRV.Elem()
		expectRT = expectRT.Elem()
	}
	if actualRV.Kind() == reflect.Ptr {
		actualRV = actualRV.Elem()
	}
	assert.Equal(t, expectRV.NumField(), actualRV.NumField())
	for i := 0; i < expectRV.NumField(); i++ {
		expectVal, ok := expectRV.Field(i).Interface().(time.Time)
		if !ok {
			assert.True(t, reflect.DeepEqual(expectRV.Field(i).Interface(), actualRV.Field(i).Interface()))
			continue
		}

		actualVal, ok := actualRV.Field(i).Interface().(time.Time)
		assert.True(t, ok)
		if expectVal.IsZero() {
			assert.True(t, actualVal.After(expectVal))
		} else if expectRT.Field(i).Name == "UpdatedAt" {
			assert.True(t, actualVal.After(expectVal))
		} else {
			assert.Equal(t, expectVal.Format(timeFormat), actualVal.Format(timeFormat))
		}
	}
}

func SliceToMap(in interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	rv := reflect.ValueOf(in)
	for i := 0; i < rv.Len(); i++ {
		val := rv.Index(i)
		isPtr := false
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
			isPtr = true
		}
		for j := 0; j < val.NumField(); j++ {
			if val.Type().Field(j).Name == "ID" {
				key := ""
				switch v := val.Field(j).Interface().(type) {
				case string:
					key = v
				case fmt.Stringer:
					key = v.String()
				default:
					panic(fmt.Sprintf("unknown %v", val))
				}
				if isPtr {
					m[key] = val.Addr().Interface()
				} else {
					m[key] = val.Interface()
				}
			}
		}
	}
	return m
}
