package cli

import (
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/ipfs-force-community/venus-common-utils/apiinfo"
	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/venus-messager/api/client"
	"github.com/filecoin-project/venus-messager/config"
)

func getAPI(ctx *cli.Context) (client.IMessager, jsonrpc.ClientCloser, error) {
	cfg, err := config.ReadConfig(ctx.String("config"))
	if err != nil {
		return &client.Message{}, func() {}, err
	}

	apiInfo := apiinfo.NewAPIInfo(cfg.API.Address, cfg.Local.Token)
	addr, err := apiInfo.DialArgs("v0")
	if err != nil {
		return nil, nil, err
	}

	client, closer, err := client.NewMessageRPC(ctx.Context, addr, apiInfo.AuthHeader())

	return client, closer, err
}
