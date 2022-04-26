package service

import (
	"context"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/venus-messager/config"

	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/ipfs-force-community/venus-common-utils/apiinfo"
)

func NewNodeClient(ctx context.Context, cfg *config.NodeConfig) (v1.FullNode, jsonrpc.ClientCloser, error) {

	apiInfo := apiinfo.NewAPIInfo(cfg.Url, cfg.Token)
	addr, err := apiInfo.DialArgs("v0")
	if err != nil {
		return nil, nil, err
	}
	res, closer, err := v1.NewFullNodeRPC(ctx, addr, apiInfo.AuthHeader())
	return res, closer, err
}
