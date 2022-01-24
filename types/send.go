package types

import (
	"github.com/filecoin-project/venus/venus-shared/types/messager"
)

const (
	ParamsJSON = messager.QuickSendParamsCodecJSON
	ParamsHex  = messager.QuickSendParamsCodecHex
)

type SendParams = messager.QuickSendParams
