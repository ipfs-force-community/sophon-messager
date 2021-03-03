package types

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/pkg/types"
)

type SendSpec struct {
	ExpireEpoch       abi.ChainEpoch `json:"expireEpoch"`
	MaxFee            types.FIL      `json:"maxFee"`
	MaxFeeCap         types.FIL      `json:"maxFeeCap"`
	GasOverEstimation float64        `json:"gasOverEstimation"`
}
