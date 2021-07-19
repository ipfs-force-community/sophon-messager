package types

import (
	"github.com/filecoin-project/go-state-types/big"
)

type SharedParams struct {
	ID uint `json:"id"`

	GasOverEstimation float64 `json:"gasOverEstimation"`
	MaxFee            big.Int `json:"maxFee,omitempty"`
	MaxFeeCap         big.Int `json:"maxFeeCap"`

	SelMsgNum uint64 `json:"selMsgNum"`
}

func (sp *SharedParams) GetMsgMeta() *MsgMeta {
	if sp == nil {
		return nil
	}
	return &MsgMeta{
		GasOverEstimation: sp.GasOverEstimation,
		MaxFee:            sp.MaxFee,
		MaxFeeCap:         sp.MaxFeeCap,
	}
}
