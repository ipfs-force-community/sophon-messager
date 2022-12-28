package testhelper

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/venus/venus-shared/testutil"
	"github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/stretchr/testify/assert"
)

// ResolveIDAddr convert a ID address to a BLS address
func ResolveIDAddr(addr address.Address) (address.Address, error) {
	if addr.Protocol() != address.ID {
		return address.Undef, fmt.Errorf("not ID address %d", addr.Protocol())
	}
	data := []byte(addr.String()[2:])
	for i := len(data); i < address.BlsPublicKeyBytes; i++ {
		data = append(data, byte(i))
	}

	return address.NewFromBytes(append([]byte{address.BLS}, data...))
}

func ResolveAddr(t *testing.T, addr address.Address) address.Address {
	if addr.Protocol() != address.ID {
		return addr
	}
	newAddr, err := ResolveIDAddr(addr)
	if err != nil {
		t.Errorf("resolve ID address failed %s %v", addr, err)
	}
	return newAddr
}

func ResolveAddrs(t *testing.T, addrs []address.Address) []address.Address {
	newAddrs := make([]address.Address, 0, len(addrs))
	for _, addr := range addrs {
		newAddrs = append(newAddrs, ResolveAddr(t, addr))
	}
	return newAddrs
}

func AddressProtocolToSignType(protocol address.Protocol) crypto.SigType {
	switch protocol {
	case address.SECP256K1:
		return crypto.SigTypeSecp256k1
	case address.BLS:
		return crypto.SigTypeBLS
	default:
		return crypto.SigTypeUnknown
	}
}

func RandAddresses(t *testing.T, count int) []address.Address {
	var addrs []address.Address
	if count < 4 {
		count = 4
	}
	for i := 0; i < count; i++ {
		if i%2 == 0 {
			addrs = append(addrs, testutil.IDAddressProvider()(t))
			continue
		}
		if i%3 == 0 {
			addrs = append(addrs, testutil.BlsAddressProvider()(t))
			continue
		}
		addrs = append(addrs, testutil.SecpAddressProvider(32)(t))
	}

	return addrs
}

const timeFormat = "2006-01-02 15:04:05.999"

func Equal(t *testing.T, expect, actual interface{}) {
	expectRV := reflect.ValueOf(expect)
	expectRT := reflect.TypeOf(expect)
	actualRV := reflect.ValueOf(actual)
	assert.Equal(t, expectRV.Kind(), actualRV.Kind())

	if expectRV.Kind() == reflect.Array || expectRT.Kind() == reflect.Slice {
		assert.Equal(t, expectRV.Len(), actualRV.Len())
		mLen := expectRV.Len()
		for i := 0; i < mLen; i++ {
			Equal(t, expectRV.Index(i).Interface(), actualRV.Index(i).Interface())
		}
		return
	}

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

func MsgGroupByAddress(msgs []*messager.Message) map[address.Address][]*messager.Message {
	addrMsgs := make(map[address.Address][]*messager.Message)
	for _, msg := range msgs {
		addrMsgs[msg.From] = append(addrMsgs[msg.From], msg)
	}

	return addrMsgs
}

func IsSortedByNonce(t *testing.T, msgs []*messager.Message) {
	addrMsgs := MsgGroupByAddress(msgs)
	for _, m := range addrMsgs {
		assert.True(t, sort.SliceIsSorted(m, func(i, j int) bool {
			return m[i].Nonce < m[j].Nonce
		}))
	}
}
