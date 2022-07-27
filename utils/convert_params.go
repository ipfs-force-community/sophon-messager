package utils

import (
	"bytes"
	"fmt"
	"math"
	"reflect"

	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/specs-actors/actors/builtin/miner"
	miner5 "github.com/filecoin-project/specs-actors/v5/actors/builtin/miner"
	market6 "github.com/filecoin-project/specs-actors/v6/actors/builtin/market"
)

var bitFieldTyp = reflect.TypeOf(bitfield.BitField{})

var poStPartitionTyp = reflect.TypeOf([]miner.PoStPartition{})

func TryConvertParams(in interface{}) (interface{}, error) {
	switch val := in.(type) {
	case *miner.ExtendSectorExpirationParams:
		out := make(map[string][]map[string]interface{}, len(val.Extensions))
		for _, extension := range val.Extensions {
			res, err := convertStructToMap(extension)
			if err != nil {
				return nil, err
			}
			out["Extensions"] = append(out["Extensions"], res)
		}
		return out, nil
	case *miner.DeclareFaultsRecoveredParams:
		out := make(map[string][]map[string]interface{}, len(val.Recoveries))
		for _, recoverie := range val.Recoveries {
			res, err := convertStructToMap(recoverie)
			if err != nil {
				return nil, err
			}
			out["Recoveries"] = append(out["Recoveries"], res)
		}
		return out, nil
	case *miner.DeclareFaultsParams:
		out := make(map[string][]map[string]interface{}, len(val.Faults))
		for _, fault := range val.Faults {
			res, err := convertStructToMap(fault)
			if err != nil {
				return nil, err
			}
			out["Faults"] = append(out["Faults"], res)
		}
		return out, nil
	case *miner5.ProveCommitAggregateParams:
		return convertStructToMap(val)
	case *miner.TerminateSectorsParams:
		out := make(map[string][]map[string]interface{}, len(val.Terminations))
		for _, termination := range val.Terminations {
			res, err := convertStructToMap(termination)
			if err != nil {
				return nil, err
			}
			out["Terminations"] = append(out["Terminations"], res)
		}
		return out, nil
	case *miner.CompactPartitionsParams:
		return convertStructToMap(val)
	case *miner.CompactSectorNumbersParams:
		return convertStructToMap(val)
	case *miner.SubmitWindowedPoStParams:
		return convertStructToMap(val)
	case *market6.PublishStorageDealsReturn:
		return convertStructToMap(val)
	}

	return in, nil
}

func convertStructToMap(in interface{}) (map[string]interface{}, error) {
	rt := reflect.TypeOf(in)
	rv := reflect.ValueOf(in)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}
	m := make(map[string]interface{}, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		if rt.Field(i).Type.AssignableTo(bitFieldTyp) {
			res, err := convertBitFieldToString(rv.Field(i).Interface().(bitfield.BitField))
			if err != nil {
				return nil, err
			}
			m[rt.Field(i).Name] = res
			continue
		}
		if rt.Field(i).Type.AssignableTo(poStPartitionTyp) {
			partitions := rv.Field(i).Interface().([]miner.PoStPartition)
			tmp := make([]map[string]interface{}, 0, len(partitions))
			for idx := range partitions {
				res, err := convertBitFieldToString(partitions[idx].Skipped)
				if err != nil {
					return nil, err
				}
				tmp = append(tmp, map[string]interface{}{
					"Index":   partitions[idx].Index,
					"Skipped": res,
				})
			}
			m[rt.Field(i).Name] = tmp
			continue
		}
		m[rt.Field(i).Name] = rv.Field(i).Interface()
	}

	return m, nil
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
