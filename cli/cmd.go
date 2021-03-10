package cli

import (
	"context"
	"net/http"

	"github.com/filecoin-project/go-jsonrpc"

	"github.com/ipfs-force-community/venus-messager/api/client"
	"github.com/ipfs-force-community/venus-messager/config"
)

func getAPI(ctx context.Context) (client.IMessager, jsonrpc.ClientCloser, error) {
	cfg, err := config.ReadConfig("./messager.toml")
	if err != nil {
		return &client.Message{}, func() {}, err
	}

	header := http.Header{}
	client, closer, err := client.NewMessageRPC(ctx, cfg.API.Address+"/rpc/v0", header)

	return client, closer, err
}
