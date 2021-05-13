package types

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/filecoin-project/go-state-types/big"
)

func init() {
	uuid, err := ParseUUID("00000000-0000-0000-0000-000000000001")
	if err != nil {
		logrus.Fatal(err)
	}
	DefGlobalFeeCfgID = uuid
}

var DefGlobalFeeCfgID UUID

type FeeConfig struct {
	ID                UUID    `json:"id"`
	WalletID          UUID    `json:"walletID"`
	Method            int64   `json:"method"`
	GasOverEstimation float64 `json:"gasOverEstimation"`
	MaxFee            big.Int `json:"maxFee"`
	MaxFeeCap         big.Int `json:"maxFeeCap"`

	IsDeleted int       `json:"isDeleted"`
	CreatedAt time.Time `json:"createAt"`
	UpdatedAt time.Time `json:"updateAt"`
}
