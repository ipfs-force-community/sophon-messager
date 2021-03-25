package cli

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
)

var ConfigCmds = &cli.Command{
	Name:  "config",
	Usage: "config commands",
	Subcommands: []*cli.Command{
		replaceCmd,
	},
}

var refreshConfigCmd = &cli.Command{
	Name:  "refresh",
	Usage: "refresh config",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		msgMeta, err := client.RefreshMsgMeta(ctx.Context)
		if err != nil {
			return err
		}

		b, err := json.MarshalIndent(msgMeta, "", "\t")
		if err != nil {
			return err
		}

		fmt.Println(b)

		return nil
	},
}
