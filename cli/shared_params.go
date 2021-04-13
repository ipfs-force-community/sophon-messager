package cli

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/ipfs-force-community/venus-messager/types"
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
	Name:      "set",
	Usage:     `set current shared params commands, eg. set: venus-messager share-params set "{\"expireEpoch\": 0, \"gasOverEstimation\": 0, \"maxFee\": 100000, \"maxFeeCap\": 20000, \"selMsgNum\": 20, \"scanInterval\": 10, \"maxEstFailNumOfMsg\": 5}"`,
	ArgsUsage: "[params]",
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() > 1 {
			return cli.ShowCommandHelp(ctx, ctx.Command.Name)
		}

		api, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if ctx.Args().Len() == 0 {
			return xerrors.Errorf("must pass params")
		}

		sp := new(types.SharedParams)
		bytes := []byte(ctx.Args().Get(0))

		err = json.Unmarshal(bytes, sp)
		if err != nil {
			return err
		}

		_, err = api.SetSharedParams(ctx.Context, sp)
		if err != nil {
			return err
		}
		fmt.Println("sp: ", *sp)

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
