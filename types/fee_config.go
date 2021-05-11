package types

import (
	"time"

	"github.com/filecoin-project/go-state-types/big"
)

type FeeConfig struct {
	ID                UUID    `json:"id"`
	WalletID          UUID    `json:"walletID"`
	MethodType        uint64  `json:"methodType"`
	GasOverEstimation float64 `json:"gasOverEstimation"`
	MaxFee            big.Int `json:"maxFee"`
	MaxFeeCap         big.Int `json:"maxFeeCap"`

	IsDeleted int       `json:"isDeleted"`
	CreatedAt time.Time `json:"createAt"`
	UpdatedAt time.Time `json:"updateAt"`
}
