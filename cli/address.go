package cli

import (
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

var walletNameFlag = &cli.StringFlag{
	Name:    "wallet-name",
	Usage:   "wallet name",
	Aliases: []string{"name"},
}

var AddrCmds = &cli.Command{
	Name:  "address",
	Usage: "address commands",
	Subcommands: []*cli.Command{
		searchAddrCmd,
		listAddrCmd,
		//deleteAddrCmd,
		//updateNonceCmd,
		forbiddenAddrCmd,
		activeAddrCmd,
		setAddrSelMsgNumCmd,
	},
}

var searchAddrCmd = &cli.Command{
	Name:      "search",
	Usage:     "search address",
	ArgsUsage: "address",
	Flags: []cli.Flag{
		walletNameFlag,
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return xerrors.Errorf("must pass address")
		}

		addr, err := address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}
		addrInfo, err := client.GetAddress(ctx.Context, ctx.String("wallet-name"), addr)
		if err != nil {
			return err
		}
		bytes, err := json.MarshalIndent(addrInfo, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var listAddrCmd = &cli.Command{
	Name:  "list",
	Usage: "list address",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		addrs, err := client.ListAddress(ctx.Context)
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

//nolint
var updateNonceCmd = &cli.Command{
	Name:      "update-nonce",
	Usage:     "update address nonce",
	ArgsUsage: "address",
	Flags: []cli.Flag{
		&cli.Uint64Flag{
			Name:  "nonce",
			Usage: "address nonce",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return xerrors.Errorf("must pass address")
		}

		addr, err := address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}

		nonce := ctx.Uint64("nonce")
		if _, err := client.UpdateNonce(ctx.Context, addr, nonce); err != nil {
			return err
		}

		return nil
	},
}

// nolint
var deleteAddrCmd = &cli.Command{
	Name:      "del",
	Usage:     "delete address",
	ArgsUsage: "address",
	Flags: []cli.Flag{
		walletNameFlag,
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return xerrors.Errorf("must pass address")
		}

		addr, err := address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}
		_, err = client.DeleteAddress(ctx.Context, ctx.String("wallet-name"), addr)
		if err != nil {
			return err
		}

		return nil
	},
}

var forbiddenAddrCmd = &cli.Command{
	Name:      "forbidden",
	Usage:     "forbidden address",
	ArgsUsage: "address",
	Flags: []cli.Flag{
		walletNameFlag,
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return xerrors.Errorf("must pass address")
		}

		addr, err := address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}
		walletName := ctx.String("wallet-name")

		hasAddr, err := client.HasAddress(ctx.Context, walletName, addr)
		if err != nil {
			return err
		}
		if !hasAddr {
			return xerrors.Errorf("address not exist")
		}

		_, err = client.ForbiddenAddress(ctx.Context, walletName, addr)
		if err != nil {
			return err
		}

		return nil
	},
}

var activeAddrCmd = &cli.Command{
	Name:      "active",
	Usage:     "activate a frozen address",
	ArgsUsage: "address",
	Flags: []cli.Flag{
		walletNameFlag,
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return xerrors.Errorf("must pass address")
		}

		addr, err := address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}
		walletName := ctx.String("wallet-name")

		hasAddr, err := client.HasAddress(ctx.Context, walletName, addr)
		if err != nil {
			return err
		}
		if !hasAddr {
			return xerrors.Errorf("address not exist")
		}

		_, err = client.ActiveAddress(ctx.Context, walletName, addr)
		if err != nil {
			return err
		}

		return nil
	},
}

var setAddrSelMsgNumCmd = &cli.Command{
	Name:      "set-sel-msg-num",
	Usage:     "set the number of address selection messages",
	ArgsUsage: "address",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "num",
			Usage: "the number of one address selection message",
		},
		walletNameFlag,
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return xerrors.Errorf("must pass address")
		}
		addr, err := address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}
		walletName := ctx.String("wallet-name")
		if _, err := client.SetSelectMsgNum(ctx.Context, walletName, addr, ctx.Uint64("num")); err != nil {
			return err
		}

		return nil
	},
}
