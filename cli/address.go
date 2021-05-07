package cli

import (
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/venus-messager/models/repo"

	"github.com/filecoin-project/go-address"

	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-messager/types"
)

var AddrCmds = &cli.Command{
	Name:  "address",
	Usage: "address commands",
	Subcommands: []*cli.Command{
		//setAddrCmd,
		searchAddrCmd,
		listAddrCmd,
		//deleteAddrCmd,
		updateNonceCmd,
	},
}

// nolint
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
		if err != nil {
			return err
		}
		addrInfo.IsDeleted = repo.NotDeleted

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

var searchAddrCmd = &cli.Command{
	Name:      "search",
	Usage:     "search address",
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

// nolint
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
