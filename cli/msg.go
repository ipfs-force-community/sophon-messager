package cli

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/ipfs-force-community/venus-messager/types"
)

var MsgCmds = &cli.Command{
	Name:  "msg",
	Usage: "msg commands",
	Subcommands: []*cli.Command{
		getCmd,
		listCmd,
	},
}

var getCmd = &cli.Command{
	Name:  "get",
	Usage: "get local msg test",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "uuid",
			Required: true,
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
		if uidStr := ctx.String("uuid"); len(uidStr) > 0 {
			uid, err := types.ParseUUID(uidStr)
			if err != nil {
				return err
			}

			msg, err = client.GetMessage(ctx.Context, uid)
			if err != nil {
				return err
			}
		} else if cidStr := ctx.String("signed_cid"); len(cidStr) > 0 {
			c, err := cid.Decode(cidStr)
			msg, err = client.GetMessageBySignedCid(ctx.Context, c.String())
			if err != nil {
				return err
			}
		} else if cidStr := ctx.String("unsigned_cid"); len(cidStr) > 0 {
			c, err := cid.Decode(cidStr)
			msg, err = client.GetMessageByUnsignedCid(ctx.Context, c.String())
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

var listCmd = &cli.Command{
	Name:  "list",
	Usage: "list local msg test",
	Flags: []cli.Flag{
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

		msg, err := client.ListMessage(ctx.Context)
		if err != nil {
			return err
		}
		bytes, err := json.MarshalIndent(msg, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var updateMessageStateCmd = &cli.Command{
	Name:  "update_msg_state",
	Usage: "manual update message state",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "id",
			Usage: "message id",
		},
		&cli.IntFlag{
			Name:     "state",
			Usage:    "message state",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "cid",
			Usage: "message unsigned cid",
		},
	},
	Action: func(ctx *cli.Context) error {
		cli, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		id := ctx.String("id")
		cid := ctx.String("cid")
		if len(id) == 0 && len(cid) == 0 {
			return xerrors.Errorf("id and cid cannot both be empty")
		}

		state := types.MessageState(ctx.Int("state"))
		if state > types.ExpireMsg {
			return xerrors.Errorf("invalid message state")
		}
		if len(id) != 0 {
			uuid, err := types.ParseUUID(id)
			if err != nil {
				return err
			}
			if _, err := cli.UpdateMessageStateByID(ctx.Context, uuid, state); err != nil {
				return err
			}
		} else if len(cid) != 0 {
			if _, err := cli.UpdateMessageStateByCid(ctx.Context, cid, state); err != nil {
				return err
			}
		}

		return nil
	},
}

var updateAllSignedMessageCmd = &cli.Command{
	Name:  "update_all_signed_msg",
	Usage: "manual update all signed message state",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "really-do-it",
			Usage: "pass this flag if you know what you are doing",
		},
	},
	Action: func(ctx *cli.Context) error {
		cli, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Bool("really-do-it") {
			return xerrors.Errorf("pass --really-do-it to confirm this action")
		}
		count, err := cli.UpdateAllSignedMessage(ctx.Context)
		if err != nil {
			return err
		}
		fmt.Printf("update message count: %d\n", count)

		return nil
	},
}

var updateSignedMessageCmd = &cli.Command{
	Name:  "update_signed_msg",
	Usage: "manual update one signed message state",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "really-do-it",
			Usage: "pass this flag if you know what you are doing",
		},
		&cli.StringFlag{
			Name:  "uuid",
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

		if !ctx.Bool("really-do-it") {
			return xerrors.Errorf("pass --really-do-it to confirm this action")
		}
		var uuid types.UUID
		if uuidStr := ctx.String("uuid"); len(uuidStr) > 0 {
			uuid, err = types.ParseUUID(uuidStr)
			if err != nil {
				return err
			}
		} else if signedCidStr := ctx.String("signed_cid"); len(signedCidStr) > 0 {
			signedCid, err := cid.Decode(signedCidStr)
			if err != nil {
				return err
			}
			msg, err := client.GetMessageBySignedCid(ctx.Context, signedCid.String())
			if err != nil {
				return err
			}
			uuid = msg.ID
		} else if unsignedCidStr := ctx.String("unsigned_cid"); len(unsignedCidStr) > 0 {
			unsignedCid, err := cid.Decode(unsignedCidStr)
			if err != nil {
				return err
			}
			msg, err := client.GetMessageByUnsignedCid(ctx.Context, unsignedCid.String())
			if err != nil {
				return err
			}
			uuid = msg.ID
		} else {
			return xerrors.Errorf("value of query must be entered")
		}

		count, err := client.UpdateSignedMessageByID(ctx.Context, uuid)
		if err != nil {
			return err
		}
		fmt.Printf("update message count: %d\n", count)

		return nil
	},
}
