package cli

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
)

var FeeConfigCmds = &cli.Command{
	Name:  "fee-config",
	Usage: "fee config commands",
	Subcommands: []*cli.Command{
		searchAddrCmd,
		listFeeConfigCmd,
	},
}

var searchFeeConfigCmd = &cli.Command{
	Name:  "search",
	Usage: "search fee config",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "wallet-name",
			Usage:    "wallet name",
			Aliases:  []string{"name"},
			Required: true,
		},
		&cli.Int64Flag{
			Name:     "method",
			Usage:    "message method",
			Required: true,
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		walletName := ctx.String("wallet-name")
		method := ctx.Int64("method")
		wallet, err := client.GetWalletByName(ctx.Context, walletName)
		if err != nil {
			return err
		}
		fcs, err := client.GetFeeConfig(ctx.Context, wallet.ID, method)
		if err != nil {
			return err
		}

		bytes, err := json.MarshalIndent(fcs, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var listFeeConfigCmd = &cli.Command{
	Name:  "list",
	Usage: "list all fee config",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		fcs, err := client.ListFeeConfig(ctx.Context)
		if err != nil {
			return err
		}

		bytes, err := json.MarshalIndent(fcs, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}
