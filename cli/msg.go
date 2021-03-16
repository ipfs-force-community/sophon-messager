package cli

import (
	"encoding/json"
	"fmt"
	"github.com/ipfs-force-community/venus-messager/types"

	"github.com/urfave/cli/v2"
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

		uidStr := ctx.String("uuid")
		uid, err := types.ParseUUID(uidStr)
		if err != nil {
			return err
		}

		msg, err := client.GetMessage(ctx.Context, uid)
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
