package cli

import (
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/venus-messager/service"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/ipfs-force-community/venus-common-utils/apiinfo"
	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/venus-messager/config"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
)

func getAPI(ctx *cli.Context) (messager.IMessager, jsonrpc.ClientCloser, error) {
	cfg, err := config.ReadConfig(ctx.String("config"))
	if err != nil {
		return nil, func() {}, err
	}

	apiInfo := apiinfo.NewAPIInfo(cfg.API.Address, cfg.JWT.Local.Token)
	addr, err := apiInfo.DialArgs("v0")
	if err != nil {
		return nil, nil, err
	}

	client, closer, err := messager.NewIMessagerRPC(ctx.Context, addr, apiInfo.AuthHeader())

	return client, closer, err
}

func getNodeAPI(ctx *cli.Context) (v1.FullNode, jsonrpc.ClientCloser, error) {
	cfg, err := config.ReadConfig(ctx.String("config"))
	if err != nil {
		return nil, func() {}, err
	}
	return service.NewNodeClient(ctx.Context, &cfg.Node)
}
