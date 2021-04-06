package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/ipfs-force-community/venus-messager/types"
)

var WalletCmds = &cli.Command{
	Name:  "wallet",
	Usage: "wallet commands",
	Subcommands: []*cli.Command{
		addWalletCmd,
		searchWalletCmd,
		listWalletCmd,
		listRemoteWalletAddrCmd,
		deleteWalletCmd,
	},
}

var addWalletCmd = &cli.Command{
	Name:  "add",
	Usage: "add a new wallet",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "wallet name",
		},
		&cli.StringFlag{
			Name:  "url",
			Usage: "wallet url",
		},
		&cli.StringFlag{
			Name:  "token",
			Usage: "wallet token",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		var w types.Wallet
		w.CreatedAt = time.Now()
		w.ID = types.NewUUID()
		w.State = types.Alive
		w.IsDeleted = -1
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

		return nil
	},
}

var searchWalletCmd = &cli.Command{
	Name:      "search",
	Usage:     "search wallet by name",
	ArgsUsage: "name",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return xerrors.Errorf("must pass name")
		}
		wallet, err := client.GetWalletByName(ctx.Context, ctx.Args().First())
		if err != nil {
			return err
		}

		bytes, err := json.MarshalIndent(wallet, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var listWalletCmd = &cli.Command{
	Name:  "list",
	Usage: "list wallet",
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

var listRemoteWalletAddrCmd = &cli.Command{
	Name:      "list-addr",
	Usage:     "list remote wallet address",
	ArgsUsage: "wallet_name",
	Aliases:   []string{"name"},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return xerrors.Errorf("must pass name")
		}

		addrs, err := client.ListRemoteWalletAddress(ctx.Context, ctx.Args().First())
		if err != nil {
			return err
		}

		bytes, err := json.MarshalIndent(addrs, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var deleteWalletCmd = &cli.Command{
	Name:      "del",
	Usage:     "delete wallet by name",
	ArgsUsage: "name",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return xerrors.Errorf("must pass name")
		}
		name := ctx.Args().First()

		_, err = client.DeleteWallet(ctx.Context, name)
		if err != nil {
			return err
		}

		return nil
	},
}
