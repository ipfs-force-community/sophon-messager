package cli

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/ipfs-force-community/venus-messager/version"
)

var VersionCmd = &cli.Command{
	Name:  "version",
	Usage: "Show venus-messager version information",
	Action: func(context *cli.Context) error {

		fmt.Println(version.GitCommit)

		return nil
	},
}
