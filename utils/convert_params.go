package utils

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"reflect"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/venus/venus-shared/types"
)

var (
	bitFieldTyp   = reflect.TypeOf(bitfield.BitField{})
	bytesTyp      = reflect.TypeOf([]byte{})
	ethAddressTyp = reflect.TypeOf(types.EthAddress{})
	addrTyp       = reflect.TypeOf(address.Address{})
)

func TryConvertParams(in interface{}) (interface{}, error) {
	rv := reflect.ValueOf(in)
	if !isNeedConvert(rv) {
		return in, nil
	}
	return convertParams(rv)
}

func isNeedConvert(v reflect.Value) bool {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.IsValid() {
		if v.Type().AssignableTo(bitFieldTyp) || v.Type().AssignableTo(ethAddressTyp) {
			return true
		}
	}
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if isNeedConvert(v.Field(i)) {
				return true
			}
		}
	case reflect.Slice:
		if v.Len() > 0 {
			if isNeedConvert(v.Index(0)) {
				return true
			}
		}
	case reflect.Map:
		if v.Len() > 0 {
			iter := v.MapRange()
			for iter.Next() {
				if isNeedConvert(iter.Value()) {
					return true
				}
			}
		}
	}
	return false
}

func convertParams(rv reflect.Value) (any, error) {
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}

	if rv.IsValid() {
		if rv.Type().AssignableTo(bitFieldTyp) {
			return convertBitFieldToString(rv.Interface().(bitfield.BitField))
		}
		if rv.Type().AssignableTo(ethAddressTyp) {
			return hexEthAddress(rv.Interface().([20]byte)), nil
		}
		if rv.Type().AssignableTo(addrTyp) {
			return rv.Interface(), nil
		}
	}
	switch rv.Kind() {
	case reflect.Slice:
		if rv.IsValid() && rv.Type().AssignableTo(bytesTyp) {
			return rv.Interface(), nil
		}
		vals := make([]interface{}, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			val, err := convertParams(rv.Index(i))
			if err != nil {
				return nil, err
			}
			vals = append(vals, val)
		}
		return vals, nil
	case reflect.Struct:
		vals := make(map[string]interface{}, rv.NumField())
		for i := 0; i < rv.NumField(); i++ {
			val, err := convertParams(rv.Field(i))
			if err != nil {
				return nil, err
			}
			vals[rv.Type().Field(i).Name] = val
		}
		return vals, nil
	case reflect.Map:
		vals := make(map[interface{}]interface{}, 0)
		iter := rv.MapRange()
		for iter.Next() {
			val, err := convertParams(iter.Value())
			if err != nil {
				return nil, err
			}
			vals[iter.Key().Interface()] = val
		}
		return vals, nil
	}
	return rv.Interface(), nil
}

func convertBitFieldToString(val bitfield.BitField) (string, error) {
	list, err := val.All(math.MaxInt64)
	if err != nil {
		return "", err
	}

	if len(list) == 0 {
		return "", nil
	}

	buf := bytes.Buffer{}
	mergeRes := merge(list)
	for i := range mergeRes {
		if len(mergeRes[i]) == 2 {
			buf.WriteString(fmt.Sprintf("%d-%d", mergeRes[i][0], mergeRes[i][1]))
		} else {
			buf.WriteString(fmt.Sprintf("%d", mergeRes[i][0]))
		}
		if i < len(mergeRes)-1 {
			buf.WriteString(", ")
		}
	}

	return buf.String(), nil
}

func merge(list []uint64) [][]uint64 {
	listLen := len(list)
	res := make([][]uint64, 0, listLen)

	start := 0
	for start < listLen {
		curr := list[start]
		end := start + 1
		for end < listLen && curr+1 == list[end] {
			curr = list[end]
			end++
		}
		if start+1 < end {
			res = append(res, []uint64{list[start], curr})
		} else {
			res = append(res, []uint64{list[start]})
		}
		start = end
	}

	return res
}

func hexEthAddress(ethAddr types.EthAddress) string {
	return "0x" + hex.EncodeToString(ethAddr[:])
}
