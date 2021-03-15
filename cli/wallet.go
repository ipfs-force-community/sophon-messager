package cli

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/ipfs-force-community/venus-messager/types"
)

var WalletCmds = &cli.Command{
	Name:  "wallet",
	Usage: "wallet commands",
	Subcommands: []*cli.Command{
		addWalletCmd,
		getWalletCmd,
		listWalletCmd,
		listWalletAddrCmd,
	},
}

var addWalletCmd = &cli.Command{
	Name:  "add",
	Usage: "add wallet",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "uuid",
			Usage:   "uuid",
			Aliases: []string{"id"},
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "name",
		},
		&cli.StringFlag{
			Name:  "url",
			Usage: "url",
		},
		&cli.StringFlag{
			Name:  "token",
			Usage: "token",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()
		var w types.Wallet

		w.ID = types.NewUUID()
		w.Name = ctx.String("name")
		w.Url = ctx.String("url")
		if len(w.Url) == 0 {
			return xerrors.Errorf("url cannot be empty")
		}
		w.Token = ctx.String("token")
		if len(w.Token) == 0 {
			return xerrors.Errorf("token cannot be empty")
		}

		_, err = client.SaveWallet(ctx.Context, &w)
		if err != nil {
			return err
		}
		fmt.Println(w)
		return nil
	},
}

var getWalletCmd = &cli.Command{
	Name:      "get",
	Usage:     "get local wallet",
	ArgsUsage: "id",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return xerrors.Errorf("must pass id")
		}
		w, err := client.GetWallet(ctx.Context, ctx.Args().First())
		if err != nil {
			return err
		}
		bytes, err := json.MarshalIndent(w, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var listWalletCmd = &cli.Command{
	Name:  "list",
	Usage: "list local wallet",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		w, err := client.ListWallet(ctx.Context)
		if err != nil {
			return err
		}

		bytes, err := json.MarshalIndent(w, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var listWalletAddrCmd = &cli.Command{
	Name:  "list-addr",
	Usage: "list local wallet",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "wallet",
			Usage: "specify which wallet to show",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		w, err := client.ListWallet(ctx.Context)
		if err != nil {
			return err
		}

		bytes, err := json.MarshalIndent(w, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}
