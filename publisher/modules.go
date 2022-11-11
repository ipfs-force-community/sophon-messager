package publisher

import (
	"context"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/models/repo"
	mpubsub "github.com/filecoin-project/venus-messager/publisher/pubsub"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/types"
	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
)

var log = logging.Logger("publisher")

func Options() fx.Option {
	return fx.Options(
		fx.Provide(NewIMsgPublisher),
		fx.Provide(newP2pPublisher),
		fx.Provide(newRpcPublisher),
	)
}

func NewIMsgPublisher(ctx context.Context, netParams *types.NetworkParams, cfg *config.PublisherConfig, P2pPublisher *P2pPublisher, rpcPublisher *RpcPublisher) (IMsgPublisher, error) {
	var ret IMsgPublisher
	var err error

	mergePublisher := NewMergePublisher(ctx, rpcPublisher)
	if cfg.EnableP2P {
		mergePublisher.AddPublisher(P2pPublisher)
	}
	ret = mergePublisher

	if cfg.Concurrency > 0 {
		ret, err = NewConcurrentPublisher(ctx, uint(cfg.Concurrency), ret)
		if err != nil {
			return nil, err
		}
	}

	if cfg.CacheReleasePeriod == 0 {
		cachePeriod := netParams.BlockDelaySecs / 3
		if cachePeriod < 1 {
			cachePeriod = 1
		}
		ret, err = NewCachePublisher(ctx, cachePeriod, ret)
		if err != nil {
			return nil, err
		}
	}
	if cfg.CacheReleasePeriod > 0 {
		cachePeriod := uint64(cfg.CacheReleasePeriod)
		ret, err = NewCachePublisher(ctx, cachePeriod, ret)
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func newP2pPublisher(pubsub mpubsub.IPubsuber, netName types.NetworkName) (*P2pPublisher, error) {
	return NewP2pPublisher(pubsub, netName)
}

func newRpcPublisher(ctx context.Context, nodeClient v1.FullNode, nodeProvider repo.INodeProvider, cfg *config.PublisherConfig) *RpcPublisher {
	if !cfg.EnableMultiNode {
		nodeProvider = nil
	}
	return NewRpcPublisher(ctx, nodeClient, nodeProvider)
}
