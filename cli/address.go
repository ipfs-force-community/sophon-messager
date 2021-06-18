package cli

import (
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

var AddrCmds = &cli.Command{
	Name:  "address",
	Usage: "address commands",
	Subcommands: []*cli.Command{
		//searchAddrCmd,
		listAddrCmd,
		deleteAddrCmd,
		//updateNonceCmd,
		forbiddenAddrCmd,
		activeAddrCmd,
		setAddrSelMsgNumCmd,
		setFeeParamsCmd,
		resetAddrCmd,
	},
}

// nolint
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

var deleteAddrCmd = &cli.Command{
	Name:      "del",
	Usage:     "delete address",
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
		fmt.Println("forbidden address success!")

		return nil
	},
}

var activeAddrCmd = &cli.Command{
	Name:      "active",
	Usage:     "activate a frozen address",
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
		fmt.Println("active address success!")

		return nil
	},
}

var setAddrSelMsgNumCmd = &cli.Command{
	Name:      "set-sel-msg-num",
	Usage:     "set the number of address selection messages",
	ArgsUsage: "address",
	Flags: []cli.Flag{
		&cli.Uint64Flag{
			Name:  "num",
			Usage: "the number of one address selection message",
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

var setFeeParamsCmd = &cli.Command{
	Name:      "set-fee-params",
	Usage:     "Address setting fee associated configuration",
	ArgsUsage: "address",
	Flags: []cli.Flag{
		&cli.Float64Flag{
			Name:  "gas-overestimation",
			Usage: "Estimate the coefficient of gas",
		},
		&cli.StringFlag{
			Name:  "max-feecap",
			Usage: "Max feecap for a message (burn and pay to miner, attoFIL/GasUnit)",
		},
		&cli.StringFlag{
			Name:  "max-fee",
			Usage: "Spend up to X attoFIL for message",
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

		_, err = client.SetFeeParams(ctx.Context, addr, ctx.Float64("gas-overestimation"), ctx.String("max-fee"), ctx.String("max-feecap"))

		return err
	},
}

var resetAddrCmd = &cli.Command{
	Name:      "reset",
	Usage:     "reset address nonce",
	ArgsUsage: "address",
	Flags: []cli.Flag{
		ReallyDoItFlag,
		&cli.Uint64Flag{
			Name:  "nonce",
			Usage: "The nonce you want to set",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Bool("really-do-it") {
			return xerrors.New("confirm to exec this command, specify --really-do-it")
		}
		if !ctx.Args().Present() {
			return xerrors.Errorf("must pass address")
		}

		addr, err := address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}
		fmt.Println("It will take dozens of seconds.")

		currentNonce, err := client.ResetAddress(ctx.Context, addr, ctx.Uint64("nonce"))
		if err != nil {
			return err
		}
		fmt.Printf("address %s current nonce %d \n", addr.String(), currentNonce)

		return nil
	},
}
