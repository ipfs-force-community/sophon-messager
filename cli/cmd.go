package cli

import (
	"net/http"
	"path/filepath"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/urfave/cli/v2"

	"github.com/ipfs-force-community/venus-messager/api/client"
	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/pkg/repo"
)

func getAPI(ctx *cli.Context) (client.IMessager, jsonrpc.ClientCloser, error) {
	path, err := repo.GetRepoPath(ctx.String("repodir"))
	if err != nil {
		return &client.Message{}, func() {}, err
	}
	cfg, err := config.ReadConfig(filepath.Join(path, repo.ConfigFilename))
	if err != nil {
		return &client.Message{}, func() {}, err
	}

	header := http.Header{}
	client, closer, err := client.NewMessageRPC(ctx.Context, cfg.API.Address+"/rpc/v0", header)

	return client, closer, err
}
