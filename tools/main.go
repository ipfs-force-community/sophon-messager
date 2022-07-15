package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/venus-messager/tools/internal"
)

func main() {
	app := &cli.App{
		Name:  "venus-messager-tools",
		Usage: "",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "The configuration file",
				Value:   "./tools_config.toml",
			},
		},
		Commands: []*cli.Command{
			internal.BatchReplaceCmd,
		},
	}

	app.Setup()

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "ERR: %v\n", err) // nolint: errcheck
	}
}
