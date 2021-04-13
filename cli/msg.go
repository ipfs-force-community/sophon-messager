package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus/pkg/constants"
	"github.com/ipfs/go-cid"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/ipfs-force-community/venus-messager/types"
)

var MsgCmds = &cli.Command{
	Name:  "msg",
	Usage: "message commands",
	Subcommands: []*cli.Command{
		searchCmd,
		listCmd,
		updateFilledMessageCmd,
		updateAllFilledMessageCmd,
		replaceCmd,
		waitMessagerCmd,
		republishCmd,
		markBadCmd,
	},
}

var searchCmd = &cli.Command{
	Name:  "search",
	Usage: "search message",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "id",
			Usage: "message id",
		},
		&cli.StringFlag{
			Name:    "signed_cid",
			Aliases: []string{"s_cid"},
			Usage:   "message signed cid",
		},
		&cli.StringFlag{
			Name:    "unsigned_cid",
			Aliases: []string{"u_cid"},
			Usage:   "message unsigned cid",
		},
		&cli.StringFlag{
			Name:  "output-type",
			Usage: "output type support json and table",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		var msg *types.Message
		if id := ctx.String("id"); len(id) > 0 {
			msg, err = client.GetMessageByUid(ctx.Context, id)
			if err != nil {
				return err
			}
		} else if cidStr := ctx.String("signed_cid"); len(cidStr) > 0 {
			c, err := cid.Decode(cidStr)
			if err != nil {
				return err
			}
			msg, err = client.GetMessageBySignedCid(ctx.Context, c)
			if err != nil {
				return err
			}
		} else if cidStr := ctx.String("unsigned_cid"); len(cidStr) > 0 {
			c, err := cid.Decode(cidStr)
			if err != nil {
				return err
			}
			msg, err = client.GetMessageByUnsignedCid(ctx.Context, c)
			if err != nil {
				return err
			}
		} else {
			return xerrors.Errorf("value of query must be entered")
		}

		bytes, err := json.MarshalIndent(msg, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var waitMessagerCmd = &cli.Command{
	Name:  "wait",
	Usage: "wait a messager msg id for result",
	Action: func(cctx *cli.Context) error {
		client, closer, err := getAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		if cctx.NArg() == 0 {
			return xerrors.New("must has id argument")
		}

		id := cctx.Args().Get(0)
		msg, err := client.WaitMessage(cctx.Context, id, uint64(constants.MessageConfidence))
		if err != nil {
			return err
		}

		fmt.Println("message cid ", msg.SignedCid)
		fmt.Println("Height:", msg.Height)
		fmt.Println("Tipset:", msg.TipSetKey.String())
		fmt.Println("exitcode:", msg.Receipt.ExitCode)
		fmt.Println("gas_used:", msg.Receipt.GasUsed)
		fmt.Println("return_value:", msg.Receipt.ReturnValue)
		return nil
	},
}

var listCmd = &cli.Command{
	Name:  "list",
	Usage: "list messages",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "output-type",
			Usage: "output type support json and table",
		},
		&cli.StringFlag{
			Name:  "from",
			Usage: "list message by address",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		var msgs []*types.Message
		if addrStr := ctx.String("from"); len(addrStr) > 0 {
			addr, err := address.NewFromString(addrStr)
			if err != nil {
				return err
			}
			msgs, err = client.ListMessageByAddress(ctx.Context, addr)
			if err != nil {
				return err
			}
		} else {
			msgs, err = client.ListMessage(ctx.Context)
			if err != nil {
				return err
			}
		}

		bytes, err := json.MarshalIndent(msgs, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var updateAllFilledMessageCmd = &cli.Command{
	Name:  "update_all_filled_msg",
	Usage: "manual update all filled message state",
	Action: func(ctx *cli.Context) error {
		cli, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		count, err := cli.UpdateAllFilledMessage(ctx.Context)
		if err != nil {
			return err
		}
		fmt.Printf("update message count: %d\n", count)

		return nil
	},
}

var updateFilledMessageCmd = &cli.Command{
	Name:  "update_filled_msg",
	Usage: "manual update one filled message state",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "id",
			Usage: "message id",
		},
		&cli.StringFlag{
			Name:    "signed_cid",
			Aliases: []string{"s_cid"},
			Usage:   "message signed cid",
		},
		&cli.StringFlag{
			Name:    "unsigned_cid",
			Aliases: []string{"u_cid"},
			Usage:   "message unsigned cid",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		var id string
		if id = ctx.String("id"); len(id) > 0 {

		} else if signedCidStr := ctx.String("signed_cid"); len(signedCidStr) > 0 {
			signedCid, err := cid.Decode(signedCidStr)
			if err != nil {
				return err
			}
			msg, err := client.GetMessageBySignedCid(ctx.Context, signedCid)
			if err != nil {
				return err
			}
			id = msg.ID
		} else if unsignedCidStr := ctx.String("unsigned_cid"); len(unsignedCidStr) > 0 {
			unsignedCid, err := cid.Decode(unsignedCidStr)
			if err != nil {
				return err
			}
			msg, err := client.GetMessageByUnsignedCid(ctx.Context, unsignedCid)
			if err != nil {
				return err
			}
			id = msg.ID
		} else {
			return xerrors.Errorf("value of query must be entered")
		}

		_, err = client.UpdateFilledMessageByID(ctx.Context, id)
		if err != nil {
			return err
		}

		return nil
	},
}

var replaceCmd = &cli.Command{
	Name:  "replace",
	Usage: "replace a message",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "gas-feecap",
			Usage: "gas feecap for new message (burn and pay to miner, attoFIL/GasUnit)",
		},
		&cli.StringFlag{
			Name:  "gas-premium",
			Usage: "gas price for new message (pay to miner, attoFIL/GasUnit)",
		},
		&cli.Int64Flag{
			Name:  "gas-limit",
			Usage: "gas limit for new message (GasUnit)",
		},
		&cli.BoolFlag{
			Name:  "auto",
			Usage: "automatically reprice the specified message",
		},
		&cli.StringFlag{
			Name:  "max-fee",
			Usage: "Spend up to X attoFIL for this message (applicable for auto mode)",
		},
	},
	ArgsUsage: "<from nonce> | <id>",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		var id string
		switch ctx.Args().Len() {
		case 1:
			id = ctx.Args().First()
		case 2:
			f, err := address.NewFromString(ctx.Args().Get(0))
			if err != nil {
				return err
			}

			n, err := strconv.ParseUint(ctx.Args().Get(1), 10, 64)
			if err != nil {
				return err
			}
			msg, err := client.GetMessageByFromAndNonce(ctx.Context, f, n)
			if err != nil {
				return fmt.Errorf("could not find referenced message: %w", err)
			}
			id = msg.ID
		default:
			return cli.ShowCommandHelp(ctx, ctx.Command.Name)
		}

		cid, err := client.ReplaceMessage(ctx.Context, id, ctx.Bool("auto"), ctx.String("max-fee"),
			ctx.Int64("gas-limit"), ctx.String("gas-premium"), ctx.String("gas-feecap"))
		if err != nil {
			return err
		}

		fmt.Println("new message cid: ", cid)
		return nil
	},
}

var republishCmd = &cli.Command{
	Name:      "republish",
	Usage:     "republish a message by id",
	ArgsUsage: "id",
	Action: func(cctx *cli.Context) error {
		client, closer, err := getAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		if cctx.NArg() == 0 {
			return xerrors.New("must has id argument")
		}

		id := cctx.Args().Get(0)
		_, err = client.RepublishMessage(cctx.Context, id)
		if err != nil {
			return err
		}
		return nil
	},
}

var markBadCmd = &cli.Command{
	Name:      "mark-bad",
	Usage:     "mark bad message",
	ArgsUsage: "id",
	Action: func(cctx *cli.Context) error {
		client, closer, err := getAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		if cctx.NArg() == 0 {
			return xerrors.New("must has id argument")
		}

		id := cctx.Args().Get(0)
		_, err = client.MarkBadMessage(cctx.Context, id)
		if err != nil {
			return err
		}
		return nil
	},
}
