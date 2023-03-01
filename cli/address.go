package cli

import (
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/urfave/cli/v2"
)

var AddrCmds = &cli.Command{
	Name:  "address",
	Usage: "address commands",
	Subcommands: []*cli.Command{
		searchAddrCmd,
		listAddrCmd,
		deleteAddrCmd,
		forbiddenAddrCmd,
		activeAddrCmd,
		setAddrSelMsgNumCmd,
		setFeeParamsCmd,
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
			return fmt.Errorf("must pass address")
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
			return fmt.Errorf("must pass address")
		}

		addr, err := address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}
		err = client.DeleteAddress(ctx.Context, addr)
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
			return fmt.Errorf("must pass address")
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
			return fmt.Errorf("address not exist")
		}

		err = client.ForbiddenAddress(ctx.Context, addr)
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
			return fmt.Errorf("must pass address")
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
			return fmt.Errorf("address not exist")
		}

		err = client.ActiveAddress(ctx.Context, addr)
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
			Name:     "num",
			Usage:    "the number of one address selection message",
			Required: true,
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return fmt.Errorf("must pass address")
		}
		addr, err := address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}
		if err := client.SetSelectMsgNum(ctx.Context, addr, ctx.Uint64("num")); err != nil {
			return err
		}

		return nil
	},
}

var setFeeParamsCmd = &cli.Command{
	Name:      "set-fee-params",
	Usage:     "Address setting fee associated configuration",
	ArgsUsage: "<address>",
	Flags: []cli.Flag{
		gasOverEstimationFlag,
		gasFeeCapFlag,
		maxFeeFlag,
		basefeeFlag,
		GasOverPremiumFlag,
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return fmt.Errorf("must pass address")
		}

		params := &messager.AddressSpec{
			GasOverEstimation: ctx.Float64(gasOverEstimationFlag.Name),
			GasOverPremium:    ctx.Float64(GasOverPremiumFlag.Name),
			MaxFeeStr:         ctx.String(maxFeeFlag.Name),
			GasFeeCapStr:      ctx.String(gasFeeCapFlag.Name),
			BaseFeeStr:        ctx.String(basefeeFlag.Name),
		}
		params.Address, err = address.NewFromString(ctx.Args().First())
		if err != nil {
			return err
		}

		return client.SetFeeParams(ctx.Context, params)
	},
}
