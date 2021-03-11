package cli

import (
	"encoding/json"
	"fmt"

	"github.com/ipfs-force-community/venus-messager/utils"
	"github.com/urfave/cli/v2"
)

var MsgCmds = &cli.Command{
	Name:  "msg",
	Usage: "msg commands",
	Subcommands: []*cli.Command{
		pushCmd,
		getCmd,
		listCmd,
	},
}

var pushCmd = &cli.Command{
	Name:  "push",
	Usage: "list local msg test",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "uuid",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()
		msg := utils.NewTestMsg()
		id, err := client.PushMessage(ctx.Context, msg)
		if err != nil {
			return err
		}
		fmt.Println(id)
		return nil
	},
}

var getCmd = &cli.Command{
	Name:  "get",
	Usage: "get local msg test",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "uuid",
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

		uid := ctx.String("uuid")
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
