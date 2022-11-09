package cli

import (
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/venus-shared/actors/builtin"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/urfave/cli/v2"

	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

var SendCmd = &cli.Command{
	Name:      "send",
	Usage:     "Send a message",
	ArgsUsage: "[targetAddress] [amount]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "from",
			Usage:    "optionally specify the address to send",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "gas-premium",
			Usage: "specify gas price to use in AttoFIL",
			Value: "0",
		},
		&cli.StringFlag{
			Name:  "gas-feecap",
			Usage: "specify gas fee cap to use in AttoFIL",
			Value: "0",
		},
		&cli.Int64Flag{
			Name:  "gas-limit",
			Usage: "specify gas limit",
			Value: 0,
		},
		&cli.Uint64Flag{
			Name:  "method",
			Usage: "specify method to invoke",
			Value: uint64(builtin.MethodSend),
		},
		&cli.StringFlag{
			Name:  "params-json",
			Usage: "specify invocation parameters in json",
		},
		&cli.StringFlag{
			Name:  "params-hex",
			Usage: "specify invocation parameters in hex",
		},
		&cli.StringFlag{
			Name:     "account",
			Usage:    "optionally specify the account to send",
			Required: false,
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() != 2 {
			return fmt.Errorf("'send' expects two arguments, target and amount")
		}

		client, close, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer close()

		var params types.QuickSendParams

		params.To, err = address.NewFromString(ctx.Args().Get(0))
		if err != nil {
			return fmt.Errorf("failed to parse target address: %w", err)
		}

		val, err := venusTypes.ParseFIL(ctx.Args().Get(1))
		if err != nil {
			return fmt.Errorf("failed to parse amount: %w", err)
		}
		params.Val = abi.TokenAmount(val)

		addr, err := address.NewFromString(ctx.String("from"))
		if err != nil {
			return fmt.Errorf("failed to parse from address: %w", err)
		}
		params.From = addr

		if ctx.IsSet("account") {
			params.Account = ctx.String("account")
		}

		if ctx.IsSet("gas-premium") {
			gp, err := venusTypes.BigFromString(ctx.String("gas-premium"))
			if err != nil {
				return err
			}
			params.GasPremium = &gp
		}

		if ctx.IsSet("gas-feecap") {
			gfc, err := venusTypes.BigFromString(ctx.String("gas-feecap"))
			if err != nil {
				return err
			}
			params.GasFeeCap = &gfc
		}

		if ctx.IsSet("gas-limit") {
			limit := ctx.Int64("gas-limit")
			params.GasLimit = &limit
		}

		params.Method = abi.MethodNum(ctx.Uint64("method"))

		if ctx.IsSet("params-json") {
			params.Params = ctx.String("params-json")
			params.ParamsType = types.QuickSendParamsCodecJSON
		}
		if ctx.IsSet("params-hex") {
			if len(params.Params) != 0 {
				return fmt.Errorf("can only specify one of 'params-json' and 'params-hex'")
			}
			params.Params = ctx.String("params-hex")
			params.ParamsType = types.QuickSendParamsCodecHex
		}

		uuid, err := client.Send(ctx.Context, params)
		if err != nil {
			return err
		}
		fmt.Printf("msg uuid %s \n", uuid)

		return nil
	},
}
