package cli

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/ipfs-force-community/venus-messager/types"
)

var NodeCmds = &cli.Command{
	Name:  "node",
	Usage: "node commands",
	Subcommands: []*cli.Command{
		addNodeCmd,
		findNodeCmd,
		listNodeCmd,
		deleteNodeCmd,
	},
}

var addNodeCmd = &cli.Command{
	Name:  "add",
	Usage: "add a new node to push message",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "node name",
		},
		&cli.StringFlag{
			Name:  "url",
			Usage: "node url",
		},
		&cli.StringFlag{
			Name:  "token",
			Usage: "node token",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		var node types.Node
		node.Name = ctx.String("name")
		node.ID = types.NewUUID()
		node.URL = ctx.String("url")
		if len(node.URL) == 0 {
			return xerrors.Errorf("url cannot be empty")
		}
		node.Token = ctx.String("token")
		if len(node.Token) == 0 {
			return xerrors.Errorf("token cannot be empty")
		}

		has, err := client.HasNode(ctx.Context, node.Name)
		if err != nil {
			return err
		}
		if has {
			return xerrors.Errorf("node exist")

		}

		_, err = client.SaveNode(ctx.Context, &node)
		if err != nil {
			return err
		}
		return nil
	},
}

var findNodeCmd = &cli.Command{
	Name:      "find",
	Usage:     "find node info",
	ArgsUsage: "name",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return xerrors.Errorf("must pass node name")
		}

		node, err := client.GetNode(ctx.Context, ctx.Args().First())
		if err != nil {
			return err
		}

		bytes, err := json.MarshalIndent(node, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var listNodeCmd = &cli.Command{
	Name:  "list",
	Usage: "list node",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		w, err := client.ListNode(ctx.Context)
		if err != nil {
			return err
		}

		bytes, err := json.MarshalIndent(w, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var deleteNodeCmd = &cli.Command{
	Name:      "del",
	Usage:     "delete node by name",
	ArgsUsage: "name",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return xerrors.Errorf("must pass name")
		}
		name := ctx.Args().First()

		_, err = client.DeleteNode(ctx.Context, name)
		if err != nil {
			return err
		}

		return nil
	},
}
