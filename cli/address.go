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
		forbiddenAddrCmd,
		permitAddrCmd,
		setAddrSelMsgNumCmd,
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
		var addrInfo types.Address

		if uuidStr := ctx.String("uuid"); len(uuidStr) > 0 {
			uuid, err := types.ParseUUID(uuidStr)
			if err != nil {
				return err
			}
			addrInfo.ID = uuid
		} else {
			addrInfo.ID = types.NewUUID()
		}

		addrStr := ctx.String("address")
		addr, err := address.NewFromString(addrStr)
		if err != nil {
			return err
		}
		addrInfo.Addr = addr
		addrInfo.Nonce = ctx.Uint64("nonce")
		addrInfo.WalletID, err = types.ParseUUID(ctx.String("wallet_id"))
		if err != nil {
			return err
		}
		addrInfo.IsDeleted = -1

		hasAddr, err := client.HasAddress(ctx.Context, addrInfo.Addr)
		if err != nil {
			return err
		}
		if hasAddr {
			return xerrors.Errorf("address exist")
		}

		_, err = client.SaveAddress(ctx.Context, &addrInfo)
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

		addr, err := address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}
		addrInfo, err := client.GetAddress(ctx.Context, addr)
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

		addr, err := address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}
		_, err = client.DeleteAddress(ctx.Context, addr)
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

		hasAddr, err := client.HasAddress(ctx.Context, addr)
		if err != nil {
			return err
		}
		if !hasAddr {
			return xerrors.Errorf("address not exist")
		}

		_, err = client.ForbiddenAddress(ctx.Context, addr)
		if err != nil {
			return err
		}

		return nil
	},
}

var permitAddrCmd = &cli.Command{
	Name:      "permit",
	Usage:     "permit address",
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

		addr, err := address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}

		hasAddr, err := client.HasAddress(ctx.Context, addr)
		if err != nil {
			return err
		}
		if !hasAddr {
			return xerrors.Errorf("address not exist")
		}

		_, err = client.ActiveAddress(ctx.Context, addr)
		if err != nil {
			return err
		}

		return nil
	},
}

var setAddrSelMsgNumCmd = &cli.Command{
	Name:      "set_sel_msg_num",
	Usage:     "set the number of address selection messages",
	ArgsUsage: "address",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name: "num",
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
		if _, err := client.SetSelectMsgNum(ctx.Context, addr, ctx.Uint64("num")); err != nil {
			return err
		}

		return nil
	},
}
