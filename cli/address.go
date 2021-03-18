package cli

import (
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/go-address"

	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/ipfs-force-community/venus-messager/types"
)

var AddrCmds = &cli.Command{
	Name:  "address",
	Usage: "address commands",
	Subcommands: []*cli.Command{
		setAddrCmd,
		findAddrCmd,
		listAddrCmd,
		deleteAddrCmd,
		updateNonceCmd,
	},
}

var setAddrCmd = &cli.Command{
	Name:  "set",
	Usage: "set local address",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "uuid",
		},
		&cli.StringFlag{
			Name:    "address",
			Usage:   "address",
			Aliases: []string{"a"},
		},
		&cli.StringFlag{
			Name:  "wallet_id",
			Usage: "bind ID of remote Wallet",
		},
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
		var addr types.Address

		if uuidStr := ctx.String("uuid"); len(uuidStr) > 0 {
			uuid, err := types.ParseUUID(uuidStr)
			if err != nil {
				return err
			}
			addr.ID = uuid
		} else {
			addr.ID = types.NewUUID()
		}

		addr.Addr = ctx.String("address")
		if len(addr.Addr) == 0 {
			return xerrors.Errorf("Address cannot be empty")
		}
		addr.Nonce = ctx.Uint64("nonce")
		addr.WalletID, err = types.ParseUUID(ctx.String("wallet_id"))
		if err != nil {
			return err
		}
		addr.IsDeleted = -1

		a, err := address.NewFromString(addr.Addr)
		if err != nil {
			return err
		}
		hasAddr, err := client.HasAddress(ctx.Context, a)
		if err != nil {
			return err
		}
		if hasAddr {
			return xerrors.Errorf("The same address exists")
		}

		_, err = client.SaveAddress(ctx.Context, &addr)
		if err != nil {
			return err
		}

		return nil
	},
}

var findAddrCmd = &cli.Command{
	Name:      "find",
	Usage:     "find local address",
	ArgsUsage: "address",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return xerrors.Errorf("must pass address")
		}

		addr, err := client.GetAddress(ctx.Context, ctx.Args().First())
		if err != nil {
			return err
		}
		bytes, err := json.MarshalIndent(addr, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var listAddrCmd = &cli.Command{
	Name:  "list",
	Usage: "list local address",
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

var updateNonceCmd = &cli.Command{
	Name:      "update_nonce",
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

		addr, err := client.GetAddress(ctx.Context, ctx.Args().First())
		if err != nil {
			return err
		}

		nonce := ctx.Uint64("nonce")
		if _, err := client.UpdateNonce(ctx.Context, addr.ID, nonce); err != nil {
			return err
		}

		return nil
	},
}

var deleteAddrCmd = &cli.Command{
	Name:      "del",
	Usage:     "delete local address",
	ArgsUsage: "address",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return xerrors.Errorf("must pass address")
		}

		_, err = client.DeleteAddress(ctx.Context, ctx.Args().First())
		if err != nil {
			return err
		}

		return nil
	},
}
