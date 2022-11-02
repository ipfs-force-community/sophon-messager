package cli

import (
	"fmt"

	"github.com/filecoin-project/go-state-types/big"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/urfave/cli/v2"
)

var GasOverPremiumFlag = &cli.Float64Flag{
	Name:  "gas-over-premium",
	Usage: "",
}

func ParseFlagToReplaceMessaeParams(cctx *cli.Context) (*messager.ReplacMessageParams, error) {
	params := messager.ReplacMessageParams{
		Auto:           cctx.Bool("auto"),
		GasLimit:       cctx.Int64("gas-limit"),
		GasOverPremium: cctx.Float64(GasOverPremiumFlag.Name),
	}

	if cctx.IsSet("max-fee") {
		maxFee, err := venusTypes.ParseFIL(cctx.String("max-fee"))
		if err != nil {
			return nil, fmt.Errorf("parse max fee failed: %v", err)
		}
		params.MaxFee = big.Int(maxFee)
	}
	if cctx.IsSet("gas-premium") {
		gasPremium, err := venusTypes.BigFromString(cctx.String("gas-premium"))
		if err != nil {
			return nil, fmt.Errorf("parse gas premium failed: %v", err)
		}
		params.GasPremium = gasPremium
	}
	if cctx.IsSet("gas-feecap") {
		gasFeecap, err := venusTypes.BigFromString(cctx.String("gas-feecap"))
		if err != nil {
			return nil, fmt.Errorf("parse gas feecap failed: %v", err)
		}
		params.GasFeecap = gasFeecap
	}

	return &params, nil
}

var reallyDoItFlag = &cli.BoolFlag{
	Name:  "really-do-it",
	Usage: "specify this flag to confirm mark-bad",
}

var outputTypeFlag = &cli.StringFlag{
	Name:  "output-type",
	Usage: "output type support json and table (default table)",
	Value: "table",
}

var FromFlag = &cli.StringFlag{
	Name:  "from",
	Usage: "address to send message",
}

var verboseFlag = &cli.BoolFlag{
	Name:  "verbose",
	Usage: "verbose address",
	Value: false,
}
