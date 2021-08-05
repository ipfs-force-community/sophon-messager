package cli

import (
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/go-state-types/big"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

var SharedParamsCmds = &cli.Command{
	Name:  "share-params",
	Usage: "share params cmd",
	Subcommands: []*cli.Command{
		setSharedParamsCmd,
		getSharedParamCmd,
		refreshSharedParamCmd,
	},
}

var setSharedParamsCmd = &cli.Command{
	Name:  "set",
	Usage: "set current shared params",
	Flags: []cli.Flag{
		&cli.Float64Flag{
			Name:  "gas-over-estimation",
			Value: 1.25,
		},
		&cli.StringFlag{
			Name:  "max-fee",
			Value: "7000000000000000",
		},
		&cli.StringFlag{
			Name:  "max-feecap",
			Value: "0",
		},
		&cli.Uint64Flag{
			Name:  "sel-msg-num",
			Value: 20,
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() > 1 {
			return cli.ShowCommandHelp(ctx, ctx.Command.Name)
		}

		api, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		params, err := api.GetSharedParams(ctx.Context)
		if err != nil {
			return err
		}
		if ctx.IsSet("gas-over-estimation") {
			params.GasOverEstimation = ctx.Float64("gas-over-estimation")
		}
		if ctx.IsSet("max-fee") {
			params.MaxFee, err = big.FromString(ctx.String("max-fee"))
			if err != nil {
				return xerrors.Errorf("parse max-fee failed %v", err)
			}
		}
		if ctx.IsSet("max-feecap") {
			params.MaxFeeCap, err = big.FromString(ctx.String("max-feecap"))
			if err != nil {
				return xerrors.Errorf("parse max-feecap failed %v", err)
			}
		}
		if ctx.IsSet("sel-msg-num") {
			params.SelMsgNum = ctx.Uint64("sel-msg-num")
		}

		_, err = api.SetSharedParams(ctx.Context, params)
		if err != nil {
			return err
		}

		return nil
	},
}

var getSharedParamCmd = &cli.Command{
	Name:  "get",
	Usage: "get shared params",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		params, err := client.GetSharedParams(ctx.Context)
		if err != nil {
			return err
		}
		paramsByte, err := json.MarshalIndent(params, "", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(paramsByte))

		return nil
	},
}

var refreshSharedParamCmd = &cli.Command{
	Name:  "refresh",
	Usage: "refresh shared params from DB",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		_, err = client.RefreshSharedParams(ctx.Context)
		if err != nil {
			return err
		}

		return nil
	},
}
