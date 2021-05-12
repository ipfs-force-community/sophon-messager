package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/filecoin-project/venus-messager/types"

	"github.com/filecoin-project/go-address"
	"golang.org/x/xerrors"

	"github.com/urfave/cli/v2"
)

var WalletAddrCmds = &cli.Command{
	Name:  "wallet-addr",
	Usage: "wallet address commands",
	Subcommands: []*cli.Command{
		searchWalletAddrCmd,
		listWalletAddrCmd,
		forbiddenAddrCmd,
		activeAddrCmd,
		setAddrSelMsgNumCmd,
	},
}

var searchWalletAddrCmd = &cli.Command{
	Name:  "search",
	Usage: "search wallet address info",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "wallet_name",
			Usage:   "wallet name",
			Aliases: []string{"name"},
		},
		&cli.StringFlag{
			Name:  "addr",
			Usage: "wallet address",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		addr, err := address.NewFromString(ctx.String("addr"))
		if err != nil {
			return err
		}
		wa, err := client.GetWalletAddress(ctx.Context, ctx.String("wallet_name"), addr)
		if err != nil {
			return err
		}

		bytes, err := json.MarshalIndent(transformWalletAddress(wa), " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var listWalletAddrCmd = &cli.Command{
	Name:  "list",
	Usage: "list wallet address",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		walletAddrs, err := client.ListWalletAddress(ctx.Context)
		if err != nil {
			return err
		}
		waFmt := make([]*walletAddressFormat, len(walletAddrs))
		for i, wa := range walletAddrs {
			waFmt[i] = transformWalletAddress(wa)
		}

		bytes, err := json.MarshalIndent(waFmt, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var forbiddenAddrCmd = &cli.Command{
	Name:      "forbidden",
	Usage:     "forbidden address",
	ArgsUsage: "address",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "wallet_name",
			Usage:   "wallet name",
			Aliases: []string{"name"},
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
		walletName := ctx.String("wallet_name")
		if len(walletName) == 0 {
			return xerrors.Errorf("must pass wallet name")
		}

		hasAddr, err := client.HasWalletAddress(ctx.Context, walletName, addr)
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
		&cli.StringFlag{
			Name:    "wallet_name",
			Usage:   "wallet name",
			Aliases: []string{"name"},
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
		walletName := ctx.String("wallet_name")

		hasAddr, err := client.HasWalletAddress(ctx.Context, walletName, addr)
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
	Name:      "set_sel_msg_num",
	Usage:     "set the number of address selection messages",
	ArgsUsage: "address",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "num",
			Usage: "the number of one address selection message",
		},
		&cli.StringFlag{
			Name:    "wallet_name",
			Usage:   "wallet name",
			Aliases: []string{"name"},
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
		walletName := ctx.String("wallet_name")
		if _, err := client.SetSelectMsgNum(ctx.Context, walletName, addr, ctx.Uint64("num")); err != nil {
			return err
		}

		return nil
	},
}

type walletAddressFormat struct {
	ID           types.UUID
	WalletID     types.UUID
	AddrID       types.UUID
	AddressState string
	SelMsgNum    uint64

	IsDeleted int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func transformWalletAddress(wa *types.WalletAddress) *walletAddressFormat {
	if wa == nil {
		return nil
	}

	return &walletAddressFormat{
		ID:           wa.ID,
		WalletID:     wa.WalletID,
		AddrID:       wa.AddrID,
		AddressState: types.StateToString(wa.AddressState),
		SelMsgNum:    wa.SelMsgNum,
		IsDeleted:    wa.IsDeleted,
		CreatedAt:    wa.CreatedAt,
		UpdatedAt:    wa.UpdatedAt,
	}
}
