package cli

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/venus-messager/service"
	"github.com/filecoin-project/venus-messager/utils/actor_parser"
	"github.com/filecoin-project/venus/venus-shared/types"
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

	apiInfo := apiinfo.NewAPIInfo(cfg.API.Address, cfg.JWT.Local.Token)
	addr, err := apiInfo.DialArgs("v0")
	if err != nil {
		return nil, nil, err
	}

	client, closer, err := client.NewMessageRPC(ctx.Context, addr, apiInfo.AuthHeader())

	return client, closer, err
}

func getNodeAPI(ctx *cli.Context) (*service.NodeClient, jsonrpc.ClientCloser, error) {
	cfg, err := config.ReadConfig(ctx.String("config"))
	if err != nil {
		return &service.NodeClient{}, func() {}, err
	}
	return service.NewNodeClient(ctx.Context, &cfg.Node)
}

func getActorGetter(ctx *cli.Context) (actor_parser.ActorGetter, jsonrpc.ClientCloser, error) {
	nodeAPI, closer, err := getNodeAPI(ctx)
	if err != nil {
		return nil, nil, err
	}
	return &getter{NodeClient: nodeAPI}, closer, nil
}

type getter struct {
	*service.NodeClient
}

func (g *getter) StateGetActor(ctx context.Context, addr address.Address, tsKey types.TipSetKey) (*types.Actor, error) {
	return g.NodeClient.StateGetActor(ctx, addr, tsKey)
}

func (g *getter) StateLookupID(ctx context.Context, addr address.Address, tsKey types.TipSetKey) (address.Address, error) {
	return g.NodeClient.StateLookupID(ctx, addr, tsKey)
}
