package cli

import (
	"errors"

	"github.com/urfave/cli/v2"
)

var LogCmds = &cli.Command{
	Name:  "log",
	Usage: "log commands",
	Subcommands: []*cli.Command{
		setLevelCmd,
	},
}

var setLevelCmd = &cli.Command{
	Name:      "set-level",
	Usage:     "set log level, eg. trace,debug,info,warn|warning,error,fatal,panic",
	ArgsUsage: "level",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		if ctx.NArg() == 0 {
			return errors.New("must has level argument")
		}

		err = client.SetLogLevel(ctx.Context, ctx.Args().Get(0))
		if err != nil {
			return err
		}
		return nil
	},
}
