package cli

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

var LogCmds = &cli.Command{
	Name:  "log",
	Usage: "log commands",
	Subcommands: []*cli.Command{
		setLevelCmd,
		logListCmd,
	},
}

var setLevelCmd = &cli.Command{
	Name:  "set-level",
	Usage: "Set the logging level.",
	UsageText: `Set the log level for logging systems:

	The system flag can be specified multiple times.

	eg) log set-level --system chain --system pubsub debug

	Available Levels:
	debug
	info
	warn
	error
 `,
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "system",
			Usage: "The system logging identifier",
		},
	},
	ArgsUsage: "<log-level>",
	Action: func(cctx *cli.Context) error {
		api, closer, err := getAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		level := strings.ToLower(cctx.Args().First())
		for _, subsystem := range cctx.StringSlice("system") {
			if err := api.SetLogLevel(cctx.Context, subsystem, level); err != nil {
				return err
			}
		}

		return nil
	},
}

var logListCmd = &cli.Command{
	Name:  "list",
	Usage: "List the logging subsystems.",
	Action: func(cctx *cli.Context) error {
		api, closer, err := getAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		subSystems, err := api.LogList(cctx.Context)
		if err != nil {
			return err
		}

		for _, s := range subSystems {
			fmt.Println(s)
		}
		return nil
	},
}
