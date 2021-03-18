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
		findWalletCmd,
		listWalletCmd,
		listRemoteWalletAddrCmd,
		deleteWalletCmd,
	},
}

var addWalletCmd = &cli.Command{
	Name:  "add",
	Usage: "add wallet",
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

var findWalletCmd = &cli.Command{
	Name:  "find",
	Usage: "find local wallet",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "uuid",
			Usage: "Search wallet according to uuid",
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "Search wallet according to name",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		var wallet *types.Wallet
		if uuidStr := ctx.String("uuid"); len(uuidStr) > 0 {
			uuid, err := types.ParseUUID(uuidStr)
			if err != nil {
				return err
			}
			wallet, err = client.GetWalletByID(ctx.Context, uuid)
			if err != nil {
				return err
			}
		} else if name := ctx.String("name"); len(name) > 0 {
			fmt.Println("name", name)
			wallet, err = client.GetWalletByName(ctx.Context, name)
			if err != nil {
				return err
			}
		} else {
			return xerrors.Errorf("value of query must be entered")
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

var listRemoteWalletAddrCmd = &cli.Command{
	Name:  "list-addr",
	Usage: "list remote wallet address by uuid",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "uuid",
			Usage: "Search data according to uuid",
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "Search data according to name",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		var uuid types.UUID
		uuidStr := ctx.String("uuid")
		if len(uuidStr) > 0 {
			uuid, err = types.ParseUUID(uuidStr)
			if err != nil {
				return err
			}
		} else if name := ctx.String("name"); len(name) > 0 {
			w, err := client.GetWalletByName(ctx.Context, name)
			if err != nil {
				return err
			}
			uuid = w.ID
		} else {
			return xerrors.Errorf("value of query must be entered")
		}

		addrs, err := client.ListRemoteWalletAddress(ctx.Context, uuid)
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
	Name:  "list-addr",
	Usage: "list remote wallet address by uuid",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "uuid",
			Usage: "Search data according to uuid",
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "Search data according to name",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		var uuid types.UUID
		uuidStr := ctx.String("uuid")
		if len(uuidStr) > 0 {
			uuid, err = types.ParseUUID(uuidStr)
			if err != nil {
				return err
			}
		} else if name := ctx.String("name"); len(name) > 0 {
			w, err := client.GetWalletByName(ctx.Context, name)
			if err != nil {
				return err
			}
			uuid = w.ID
		} else {
			return xerrors.Errorf("value of query must be entered")
		}

		addrs, err := client.ListRemoteWalletAddress(ctx.Context, uuid)
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
