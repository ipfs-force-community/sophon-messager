package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/filecoin-project/venus-messager/utils/actor_parser"
	"strconv"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/go-state-types/exitcode"
	"github.com/filecoin-project/venus/pkg/constants"
	venusTypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs/go-cid"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-messager/cli/tablewriter"
	"github.com/filecoin-project/venus-messager/types"
)

var ReallyDoItFlag = &cli.BoolFlag{
	Name:  "really-do-it",
	Usage: "specify this flag to confirm mark-bad",
}

var MsgCmds = &cli.Command{
	Name:  "msg",
	Usage: "message commands",
	Subcommands: []*cli.Command{
		searchCmd,
		listCmd,
		listFailedCmd,
		ListBlockedMessageCmd,
		updateFilledMessageCmd,
		updateAllFilledMessageCmd,
		replaceCmd,
		waitMessagerCmd,
		republishCmd,
		markBadCmd,
		clearUnFillMessageCmd,
		recoverFailedMsgCmd,
	},
}

var outputTypeFlag = &cli.StringFlag{
	Name:  "output-type",
	Usage: "output type support json and table (default table)",
	Value: "table",
}

var FromFlag = &cli.StringFlag{
	Name:  "from",
	Usage: "address to send message",
}

var verboseFlag = &cli.BoolFlag{
	Name:  "verbose",
	Usage: "verbose address",
	Value: false,
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
			Name:  "cid",
			Usage: "message cid",
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
		} else if cidStr := ctx.String("cid"); len(cidStr) > 0 {
			c, err := cid.Decode(cidStr)
			if err != nil {
				return err
			}
			msg, err = client.GetMessageBySignedCid(ctx.Context, c)
			if err != nil {
				return err
			}
		} else {
			return xerrors.Errorf("value of query must be entered")
		}

		bytes, err := json.MarshalIndent(transformMessage(msg), "", "\t")
		if err != nil {
			return err
		}
		fmt.Printf("- message information:\n%s\n", string(bytes))

		getter, closer, err := getActorGetter(ctx)
		if err != nil {
			return nil
		}
		defer closer()
		paser, err := actor_parser.NewMessageParser(getter)
		if err != nil {
			return nil
		}
		params, rets, err := paser.ParseMessage(ctx.Context, msg.VMMessage(), msg.Receipt)
		if err != nil {
			return err
		}

		if params != nil {
			bytes, _ := json.MarshalIndent(params, "", "\t")
			if len(bytes) != 0 {
				fmt.Printf("- params information:\n%s\n", string(bytes))
			}
		}

		if rets != nil {
			bytes, _ := json.MarshalIndent(rets, "", "\t")
			if len(bytes) != 0 {
				fmt.Printf("- returns information:\n%s\n", string(bytes))
			}
		}

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
		fmt.Println("return_value:", string(msg.Receipt.ReturnValue))
		return nil
	},
}

var listCmd = &cli.Command{
	Name:  "list",
	Usage: "list messages",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "page-index",
			Usage: "pagination index, start from 1",
			Value: 1,
		},
		&cli.IntFlag{
			Name:  "page-size",
			Usage: "pagination size, default tob 100",
			Value: 100,
		},
		FromFlag,
		outputTypeFlag,
		verboseFlag,
		&cli.IntFlag{
			Name:  "state",
			Value: int(types.UnFillMsg),
			Usage: `filter by message state,
state:
  1:  UnFillMsg
  2:  FillMsg
  3:  OnChainMsg
  4:  FailedMsg
  5:  ReplacedMsg
  6:  NoWalletMsg
`,
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()
		var from address.Address
		if addrStr := ctx.String("from"); len(addrStr) > 0 {
			from, err = address.NewFromString(addrStr)
			if err != nil {
				return err
			}
		}

		state := types.MessageState(ctx.Int("state"))

		pageIndex := ctx.Int("page-index")
		pageSize := ctx.Int("page-size")

		msgs, err := client.ListMessageByFromState(ctx.Context, from, state, false, pageIndex, pageSize)
		if err != nil {
			return err
		}

		if ctx.String("output-type") == "table" {
			return outputWithTable(msgs, ctx.Bool("verbose"))
		}
		msgT := make([]*message, 0, len(msgs))
		for _, msg := range msgs {
			msgT = append(msgT, transformMessage(msg))
		}
		bytes, err := json.MarshalIndent(msgT, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))

		return nil
	},
}

var listFailedCmd = &cli.Command{
	Name:  "list-fail",
	Usage: "list failed messages",
	Flags: []cli.Flag{
		FromFlag,
		outputTypeFlag,
		verboseFlag,
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		var msgs []*types.Message

		msgs, err = client.ListFailedMessage(ctx.Context)
		if err != nil {
			return err
		}

		if addrStr := ctx.String("from"); len(addrStr) > 0 {
			newMsgs := make([]*types.Message, 0, len(msgs))
			for _, msg := range msgs {
				if msg.From.String() == addrStr {
					newMsgs = append(newMsgs, msg)
				}
			}
			msgs = newMsgs
		}

		if ctx.String("output-type") == "table" {
			return outputWithTable(msgs, ctx.Bool("verbose"))
		}
		msgT := make([]*message, 0, len(msgs))
		for _, msg := range msgs {
			msgT = append(msgT, transformMessage(msg))
		}
		bytes, err := json.MarshalIndent(msgT, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))

		return nil
	},
}

var ListBlockedMessageCmd = &cli.Command{
	Name:  "list-blocked",
	Usage: "Lists messages that have not been chained for a period of time",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "time",
			Usage:   "exceeding residence time, eg. 3s,3m,3h (default 3h)",
			Aliases: []string{"t"},
			Value:   "3h",
		},
		FromFlag,
		outputTypeFlag,
		verboseFlag,
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		var msgs []*types.Message
		var addr address.Address

		t, err := time.ParseDuration(ctx.String("time"))
		if err != nil {
			return err
		}
		if ctx.IsSet("from") {
			addr, err = address.NewFromString(ctx.String("from"))
			if err != nil {
				return err
			}
		}

		msgs, err = client.ListBlockedMessage(ctx.Context, addr, t)
		if err != nil {
			return err
		}

		if ctx.String("output-type") == "table" {
			return outputWithTable(msgs, ctx.Bool("verbose"))
		}
		msgT := make([]*message, 0, len(msgs))
		for _, msg := range msgs {
			msgT = append(msgT, transformMessage(msg))
		}
		bytes, err := json.MarshalIndent(msgT, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))

		return nil
	},
}

var tw = tablewriter.New(
	tablewriter.Col("ID"),
	tablewriter.Col("To"),
	tablewriter.Col("From"),
	tablewriter.Col("Nonce"),
	tablewriter.Col("Value"),
	tablewriter.Col("GasLimit"),
	tablewriter.Col("GasFeeCap"),
	tablewriter.Col("GasPremium"),
	tablewriter.Col("Method"),
	tablewriter.Col("State"),
	tablewriter.Col("ExitCode"),
	tablewriter.Col("CreateAt"),
)

func outputWithTable(msgs []*types.Message, verbose bool) error {
	for _, msgT := range msgs {
		msg := transformMessage(msgT)
		val := venusTypes.MustParseFIL(msg.UnsignedMessage.Value.String() + "attofil").String()
		row := map[string]interface{}{
			"ID":         msg.ID,
			"To":         msg.UnsignedMessage.To,
			"From":       msg.UnsignedMessage.From,
			"Nonce":      msg.UnsignedMessage.Nonce,
			"Value":      val,
			"GasLimit":   msg.UnsignedMessage.GasLimit,
			"GasFeeCap":  msg.UnsignedMessage.GasFeeCap,
			"GasPremium": msg.UnsignedMessage.GasPremium,
			"Method":     msg.UnsignedMessage.Method,
			"State":      msg.State,
			"CreateAt":   msg.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if !verbose {
			if from := msg.UnsignedMessage.From.String(); len(from) > 9 {
				row["From"] = from[:9] + "..."
			}
			if to := msg.UnsignedMessage.To.String(); len(to) > 9 {
				row["To"] = to[:9] + "..."
			}
			if len(msg.ID) > 36 {
				row["ID"] = msg.ID[:36] + "..."
			}
			if len(val) > 6 {
				row["Value"] = val[:6] + "..."
			}
		}
		if msg.Receipt != nil {
			row["ExitCode"] = msg.Receipt.ExitCode
		}
		tw.Write(row)
	}

	buf := new(bytes.Buffer)
	if err := tw.Flush(buf); err != nil {
		return err
	}
	fmt.Println(buf)
	return nil
}

var updateAllFilledMessageCmd = &cli.Command{
	Name:  "update-all-filled-msg",
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
	Name:  "update-filled-msg",
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
	Name:  "mark-bad",
	Usage: "mark bad message",
	Flags: []cli.Flag{
		ReallyDoItFlag,
		&cli.StringFlag{
			Name:  "from",
			Usage: "mark unfill message as bad message if specify this flag",
		},
	},
	ArgsUsage: "id slice",
	Action: func(cctx *cli.Context) error {
		client, closer, err := getAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		if !cctx.Bool("really-do-it") {
			return errors.New("confirm to exec this command, specify --really-do-it")
		}

		if cctx.IsSet("from") {
			fromAddr, err := address.NewFromString(cctx.String("from"))
			if err != nil {
				return err
			}
			msgs, err := client.ListMessageByAddress(cctx.Context, fromAddr)
			if err != nil {
				return err
			}
			for _, msg := range msgs {
				if msg.State == types.UnFillMsg {
					if msg.Receipt != nil && len(msg.Receipt.ReturnValue) > 0 {
						_, err = client.MarkBadMessage(cctx.Context, msg.ID)
						if err != nil {
							fmt.Printf("mark msg %s as bad failed: %v\n", msg.ID, err)
							continue
						}
						fmt.Printf("mark msg %s as bad successful\n", msg.ID)
					}
				}
			}
		} else {
			if cctx.NArg() == 0 {
				return xerrors.New("must has id argument")
			}
			for _, id := range cctx.Args().Slice() {
				_, err = client.MarkBadMessage(cctx.Context, id)
				if err != nil {
					fmt.Printf("mark msg %s as bad failed: %v\n", id, err)
					continue
				}
				fmt.Printf("mark msg %s as bad successful\n", id)
			}
		}

		return nil
	},
}

var recoverFailedMsgCmd = &cli.Command{
	Name:  "recover-failed-msg",
	Usage: "recover failed messages",
	Flags: []cli.Flag{
		ReallyDoItFlag,
		&cli.StringFlag{
			Name:  "from",
			Usage: "mark unfill message as bad message if specify this flag",
		},
	},
	ArgsUsage: "id slice",
	Action: func(cctx *cli.Context) error {
		client, closer, err := getAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		if !cctx.Bool("really-do-it") {
			return errors.New("confirm to exec this command, specify --really-do-it")
		}

		if cctx.IsSet("from") {
			fromAddr, err := address.NewFromString(cctx.String("from"))
			if err != nil {
				return err
			}
			msgIDs, err := client.RecoverFailedMsg(cctx.Context, fromAddr)
			if err != nil {
				return err
			}
			fmt.Printf("recover failed messages success: %v \n", msgIDs)
		} else {
			if cctx.NArg() == 0 {
				return xerrors.New("must has id argument")
			}
			for _, id := range cctx.Args().Slice() {
				msg, err := client.GetMessageByUid(cctx.Context, id)
				if err != nil {
					return err
				}
				if msg.State != types.FailedMsg {
					fmt.Printf("need failed msg: %v", id)
					continue
				}
				if msg.Nonce != 0 && msg.Signature != nil {
					_, err = client.UpdateMessageStateByID(cctx.Context, id, types.FillMsg)
					if err != nil {
						fmt.Printf("recover msg %s failed: %v", id, err)
						continue
					}
					fmt.Printf("recover msg %s success, current state: %v, after recover msg state: %v\n", id, msg.State, types.FillMsg)
				} else {
					_, err = client.UpdateMessageStateByID(cctx.Context, id, types.UnFillMsg)
					if err != nil {
						fmt.Printf("recover msg %s failed: %v", id, err)
						continue
					}
					fmt.Printf("recover msg %s success, current state: %v, after recover msg state: %v\n", id, msg.State, types.UnFillMsg)
				}
			}
		}

		return nil
	},
}

type message struct {
	ID string

	UnsignedCid *cid.Cid
	SignedCid   *cid.Cid
	venusTypes.UnsignedMessage
	Signature *crypto.Signature

	Height     int64
	Confidence int64
	Receipt    *receipt
	TipSetKey  venusTypes.TipSetKey

	Meta *types.MsgMeta

	WalletName string
	FromUser   string

	State string

	UpdatedAt time.Time
	CreatedAt time.Time
}

type receipt struct {
	ExitCode    exitcode.ExitCode
	ReturnValue string
	GasUsed     int64
}

func transformMessage(msg *types.Message) *message {
	if msg == nil {
		return nil
	}

	m := &message{
		ID:              msg.ID,
		UnsignedCid:     msg.UnsignedCid,
		SignedCid:       msg.SignedCid,
		UnsignedMessage: msg.UnsignedMessage,
		Signature:       msg.Signature,
		Height:          msg.Height,
		Confidence:      msg.Confidence,
		TipSetKey:       msg.TipSetKey,
		Meta:            msg.Meta,
		WalletName:      msg.WalletName,
		FromUser:        msg.FromUser,
		State:           types.MsgStateToString(msg.State),
		UpdatedAt:       msg.UpdatedAt,
		CreatedAt:       msg.CreatedAt,
	}
	if msg.Receipt != nil {
		m.Receipt = &receipt{
			ExitCode:    msg.Receipt.ExitCode,
			ReturnValue: string(msg.Receipt.ReturnValue),
			GasUsed:     msg.Receipt.GasUsed,
		}
	}

	return m
}

var clearUnFillMessageCmd = &cli.Command{
	Name:      "clear-unfill-msg",
	Usage:     "clear unfill messages by address",
	ArgsUsage: "address",
	Flags: []cli.Flag{
		ReallyDoItFlag,
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

		count, err := client.ClearUnFillMessage(ctx.Context, addr)
		if err != nil {
			return err
		}
		fmt.Printf("clear %d unfill messages \n", count)

		return nil
	},
}
