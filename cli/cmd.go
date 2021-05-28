package cli

import (
	"net/http"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/venus-messager/api/client"
	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/utils"
)

func getAPI(ctx *cli.Context) (client.IMessager, jsonrpc.ClientCloser, error) {
	cfg, err := config.ReadConfig(ctx.String("config"))
	if err != nil {
		return &client.Message{}, func() {}, err
	}
	addr, err := utils.DialArgs(cfg.API.Address)
	if err != nil {
		return &client.Message{}, func() {}, err
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+cfg.Local.Token)
	client, closer, err := client.NewMessageRPC(ctx.Context, addr, header)

	return client, closer, err
}
