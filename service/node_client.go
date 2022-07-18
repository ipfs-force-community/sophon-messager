package service

import (
	"context"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/venus-messager/config"

	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
)

func NewNodeClient(ctx context.Context, cfg *config.NodeConfig) (v1.FullNode, jsonrpc.ClientCloser, error) {
	return v1.DialFullNodeRPC(ctx, cfg.Url, cfg.Token, nil)
}
