package cli

import (
	"net/http"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/urfave/cli/v2"

	"github.com/ipfs-force-community/venus-messager/api/client"
	"github.com/ipfs-force-community/venus-messager/config"
)

func getAPI(ctx *cli.Context) (client.IMessager, jsonrpc.ClientCloser, error) {
	cfg, err := config.ReadConfig(ctx.String("config"))
	if err != nil {
		return &client.Message{}, func() {}, err
	}

	header := http.Header{}
	client, closer, err := client.NewMessageRPC(ctx.Context, "http://"+cfg.API.Address+"/rpc/v0", header)

	return client, closer, err
}
