package cli

import (
	"fmt"

	"github.com/filecoin-project/venus/pkg/net"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/urfave/cli/v2"
)

var SwarmCmds = &cli.Command{
	Name:  "swarm",
	Usage: "swarm commands",
	Subcommands: []*cli.Command{
		addressListenCmd,
		connectCmd,
		peersCmd,
	},
}

var connectCmd = &cli.Command{
	Name:      "connect",
	Usage:     "connect to a libp2p node",
	ArgsUsage: "[peerIds]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "address",
			Aliases: []string{"a"},
			Usage:   "connect with libp2p address rather than peerId",
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() < 1 {
			return fmt.Errorf("must specify peerId or address")
		}

		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		peers := ctx.Args().Slice()

		if ctx.IsSet("address") {
			pis, err := net.ParseAddresses(ctx.Context, peers)
			if err != nil {
				return err
			}
			for _, pi := range pis {
				if err := client.NetConnect(ctx.Context, pi); err != nil {
					return err
				}
			}
		} else {
			for _, p := range peers {
				peerid, err := peer.Decode(p)
				if err != nil {
					return fmt.Errorf("invalid peer id: %s", p)
				}
				pi, err := client.NetFindPeer(ctx.Context, peerid)
				if err != nil {
					return err
				}
				if err := client.NetConnect(ctx.Context, pi); err != nil {
					return err
				}
			}
		}

		return nil
	},
}

var peersCmd = &cli.Command{
	Name:  "peers",
	Usage: "list swarm peers",
	Action: func(ctx *cli.Context) error {

		api, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		peers, err := api.NetPeers(ctx.Context)
		if err != nil {
			return err
		}

		for _, p := range peers {
			fmt.Printf("%s, %s\n", p.ID, p.Addrs)
		}

		return nil
	},
}

var addressListenCmd = &cli.Command{
	Name:  "listen",
	Usage: "output the listen addresses",
	Action: func(ctx *cli.Context) error {

		api, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		peerInfo, err := api.NetAddrsListen(ctx.Context)
		if err != nil {
			return err
		}

		fmt.Printf("%s, %s\n", peerInfo.ID, peerInfo.Addrs)
		return nil
	},
}
