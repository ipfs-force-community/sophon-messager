package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/ipfs-force-community/venus-messager/api/client"
	"github.com/ipfs-force-community/venus-messager/types"
)

var AddrCmds = &cli.Command{
	Name:  "address",
	Usage: "address commands",
	Subcommands: []*cli.Command{
		setAddrCmd,
		getAddrCmd,
		listAddrCmd,
	},
}

var setAddrCmd = &cli.Command{
	Name:  "set",
	Usage: "set local address",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "address",
			Usage:   "address",
			Aliases: []string{"a"},
		},
		&cli.Uint64Flag{
			Name:  "nonce",
			Usage: "the address corresponds to the nonce",
		},
	},
	Action: func(ctx *cli.Context) error {
		header := http.Header{}
		client, closer, err := client.NewMessageRPC(ctx.Context, "http://127.0.0.1:39812/rpc/v0", header)
		if err != nil {
			return err
		}
		defer closer()
		var addr types.Address

		addr.Addr = ctx.String("address")
		if len(addr.Addr) == 0 {
			return xerrors.Errorf("Address cannot be empty")
		}
		addr.Nonce = ctx.Uint64("nonce")

		_, err = client.SaveAddress(ctx.Context, &addr)
		if err != nil {
			return err
		}
		fmt.Println(addr)
		return nil
	},
}

var getAddrCmd = &cli.Command{
	Name:      "get",
	Usage:     "get local address",
	ArgsUsage: "address",
	Action: func(ctx *cli.Context) error {
		header := http.Header{}
		client, closer, err := client.NewMessageRPC(ctx.Context, "http://127.0.0.1:39812/rpc/v0", header)
		if err != nil {
			return err
		}
		defer closer()

		if !ctx.Args().Present() {
			return xerrors.Errorf("must pass address")
		}
		addr, err := client.GetAddress(ctx.Context, ctx.Args().First())
		if err != nil {
			return err
		}
		bytes, err := json.MarshalIndent(addr, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var listAddrCmd = &cli.Command{
	Name:  "list",
	Usage: "list local address",
	Action: func(ctx *cli.Context) error {
		header := http.Header{}
		client, closer, err := client.NewMessageRPC(ctx.Context, "http://127.0.0.1:39812/rpc/v0", header)
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
