package cli

import (
	"encoding/json"
	"fmt"
	"github.com/ipfs-force-community/venus-messager/api/client"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/urfave/cli/v2"
	"net/http"
	"time"
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
		header := http.Header{}
		messageClient, closer, err := client.NewMessageRPC(ctx.Context, "http://127.0.0.1:39812/rpc/v0", header)
		if err != nil {
			return err
		}
		defer closer()
		uid := ctx.String("uuid")
		msg := types.Message{
			Id:        uid,
			Version:   0,
			To:        "1111",
			From:      "33333",
			Nonce:     0,
			GasLimit:  0,
			Method:    0,
			Params:    nil,
			SignData:  nil,
			IsDeleted: -1,
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		}
		id, err := messageClient.PushMessage(ctx.Context, &msg)
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
		header := http.Header{}
		messageClient, closer, err := client.NewMessageRPC(ctx.Context, "http://127.0.0.1:39812/rpc/v0", header)
		if err != nil {
			return err
		}
		defer closer()

		uid := ctx.String("uuid")
		msg, err := messageClient.GetMessage(ctx.Context, uid)
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
		header := http.Header{}
		messageClient, closer, err := client.NewMessageRPC(ctx.Context, "http://127.0.0.1:39812/rpc/v0", header)
		if err != nil {
			return err
		}
		defer closer()

		msg, err := messageClient.ListMessage(ctx.Context)
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
