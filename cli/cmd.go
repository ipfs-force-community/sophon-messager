package cli

import (
	"path/filepath"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/venus-messager/filestore"
	"github.com/filecoin-project/venus-messager/service"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/ipfs-force-community/venus-common-utils/apiinfo"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/venus-messager/config"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
)

func getAPI(ctx *cli.Context) (messager.IMessager, jsonrpc.ClientCloser, error) {
	cfg, err := getConfig(ctx)
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
	cfg, err := getConfig(ctx)
	if err != nil {
		return nil, func() {}, err
	}
	return service.NewNodeClient(ctx.Context, &cfg.Node)
}

func getConfig(ctx *cli.Context) (*config.Config, error) {
	repoPath, err := homedir.Expand(ctx.String("repo"))
	if err != nil {
		return nil, err
	}

	return config.ReadConfig(filepath.Join(repoPath, filestore.ConfigFile))
}
