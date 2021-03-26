package cli

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/ipfs-force-community/venus-messager/types"
)

var SharedParamsCmds = &cli.Command{
	Name:      "share-params",
	Usage:     `get or set current shared params commands, eg. set: venus-messager share-params "{\"expireEpoch\": 10000, \"gasOverEstimation\": 0, \"maxFee\": 100000, \"maxFeeCap\": 20000, \"selMsgNum\": 12, \"scanInterval\": 8000000000, \"maxEstFailNumOfMsg\": 5}"`,
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
			sp, err := api.GetSharedParams(ctx.Context)
			if err != nil {
				return err
			}

			bytes, err := json.Marshal(sp)
			if err != nil {
				return err
			}

			fmt.Println(string(bytes))
		} else {
			sp := new(types.SharedParams)
			bytes := []byte(ctx.Args().Get(0))

			err := json.Unmarshal(bytes, sp)
			if err != nil {
				return err
			}

			sp, err = api.SetSharedParams(ctx.Context, sp)
			if err != nil {
				return err
			}
			fmt.Println("sp:", *sp)
		}

		return nil
	},
}
