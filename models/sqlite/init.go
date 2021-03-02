package sqlite

import (
	"fmt"
	"reflect"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/hunjixin/automapper"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/shopspring/decimal"

	"github.com/ipfs-force-community/venus-messager/types"
)

var TMessage = reflect.TypeOf(&types.Message{})
var TSqliteMessage = reflect.TypeOf(&sqliteMessage{})

var TWallet = reflect.TypeOf(&types.Wallet{})
var TSqliteWallet = reflect.TypeOf(&sqliteWallet{})

var TAddress = reflect.TypeOf(&types.Address{})
var TSqliteAddress = reflect.TypeOf(&sqliteAddress{})

var ERRUnspportedMappingType = fmt.Errorf("unsupported mapping type")

func init() {
	var nulldecimal = func(b big.Int) decimal.NullDecimal {
		return decimal.NullDecimal{decimal.NewFromBigInt(b.Int, 0), b.Int != nil}
	}
	var fromDecimal = func(decimal decimal.NullDecimal) big.Int {
		i := big.NewInt(0)
		if decimal.Valid {
			i.Int.SetString(decimal.Decimal.String(), 10)
		}
		return i
	}

	automapper.MustCreateMapper((*types.Message)(nil), (*sqliteMessage)(nil)).
		Mapping(func(destVal reflect.Value, sourceVal interface{}) error {
			var srcMsg *types.Message
			var destPoint, isok = destVal.Interface().(**sqliteMessage)
			if !isok {
				return ERRUnspportedMappingType
			}
			if srcMsg, isok = sourceVal.(*types.Message); !isok {
				return ERRUnspportedMappingType
			}
			destMsg := &sqliteMessage{}

			*destPoint = destMsg

			destMsg.GasLimit = srcMsg.GasLimit
			destMsg.Uid = srcMsg.Uid
			destMsg.Version = srcMsg.Version
			destMsg.To = srcMsg.To.String()
			destMsg.From = srcMsg.From.String()
			destMsg.Nonce = srcMsg.Nonce
			destMsg.Value = nulldecimal(srcMsg.Value)
			destMsg.GasLimit = srcMsg.GasLimit
			destMsg.GasFeeCap = nulldecimal(srcMsg.GasFeeCap)
			destMsg.GasPremium = nulldecimal(srcMsg.GasPremium)
			destMsg.Method = int(srcMsg.Method)
			destMsg.Params = srcMsg.Params
			destMsg.Signature = (*repo.SqlSignature)(srcMsg.Signature)
			destMsg.Cid = srcMsg.UnsingedCid().String()
			destMsg.SignedCid = srcMsg.SignedCid().String()
			destMsg.Meta = srcMsg.Meta
			return nil
		})

	automapper.MustCreateMapper((*sqliteMessage)(nil), (*types.Message)(nil)).
		Mapping(func(destVal reflect.Value, sourceVal interface{}) error {
			var destMsg *types.Message
			var srcMsg, isok = destVal.Interface().(*sqliteMessage)
			if !isok {
				return ERRUnspportedMappingType
			}
			if destMsg, isok = sourceVal.(*types.Message); !isok {
				return ERRUnspportedMappingType
			}
			destMsg.Uid = srcMsg.Uid
			destMsg.Version = srcMsg.Version
			destMsg.To, _ = address.NewFromString(srcMsg.To)
			destMsg.From, _ = address.NewFromString(srcMsg.From)
			destMsg.Nonce = srcMsg.Nonce
			destMsg.Value = fromDecimal(srcMsg.Value)
			destMsg.GasLimit = srcMsg.GasLimit
			destMsg.GasFeeCap = fromDecimal(srcMsg.GasFeeCap)
			destMsg.GasPremium = fromDecimal(srcMsg.GasPremium)
			destMsg.Method = abi.MethodNum(srcMsg.Method)
			destMsg.Params = srcMsg.Params
			destMsg.Signature = (*crypto.Signature)(srcMsg.Signature)
			destMsg.Meta = srcMsg.Meta
			return nil
		})
}
