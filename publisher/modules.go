package publisher

import (
	"context"
	"sort"

	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/sophon-messager/config"
	"github.com/ipfs-force-community/sophon-messager/models/repo"
	"github.com/ipfs-force-community/sophon-messager/utils"
	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
)

var log = logging.Logger("publisher")

type MessageReceiver chan []*types.SignedMessage

func Options() fx.Option {
	return fx.Options(
		fx.Provide(NewMessageReciver),
		fx.Provide(NewIMsgPublisher),
		fx.Provide(NewP2pPublisher),
		fx.Provide(newRpcPublisher),
	)
}

func NewMessageReciver(ctx context.Context, p IMsgPublisher) (MessageReceiver, error) {
	msgReceiver := make(MessageReceiver, 100)
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Infof("context done, stop receive message")
				return
			case msgs := <-msgReceiver:
				for addr, tMsgs := range utils.MsgsGroupByAddress(msgs) {
					sort.Slice(tMsgs, func(i, j int) bool {
						return tMsgs[i].Message.Nonce < tMsgs[j].Message.Nonce
					})
					if err := p.PublishMessages(ctx, tMsgs); err != nil {
						log.Warnw("publish message failed", "addr", addr.String(), "msg len", len(tMsgs), "err", err)
					}
				}
			}
		}
	}()
	return msgReceiver, nil
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

func newRpcPublisher(ctx context.Context, nodeClient v1.FullNode, nodeProvider repo.INodeProvider, cfg *config.PublisherConfig) *RpcPublisher {
	return NewRpcPublisher(ctx, nodeClient, nodeProvider, cfg.EnableMultiNode)
}
