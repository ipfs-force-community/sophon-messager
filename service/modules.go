package service

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.uber.org/fx"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/log"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/types"
)

func MessagerService() fx.Option {
	return fx.Options(
		fx.Provide(NewMessageService),
		//fx.Provide(NewWalletService),
		fx.Provide(NewAddressService),
		fx.Provide(NewSharedParamsService),
		fx.Provide(NewNodeService),
		fx.Provide(newMessagePubSubIndirect),
	)
}

func StartNodeEvents(lc fx.Lifecycle, client v1.FullNode, msgService *MessageService, log *log.Logger) *NodeEvents {
	nd := &NodeEvents{
		client:     client,
		log:        log,
		msgService: msgService,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go msgService.StartPushMessage(ctx, msgService.fsRepo.Config().MessageService.SkipPushMessage)
			go func() {
				for {
					if err := nd.listenHeadChangesOnce(ctx); err != nil {
						log.Errorf("listen head changes errored: %s", err)
					} else {
						log.Warn("listenHeadChanges quit")
					}
					select {
					case <-time.After(time.Second):
					case <-ctx.Done():
						log.Warnf("not restarting listenHeadChanges: context error: %s", ctx.Err())
						return
					}

					log.Info("restarting listenHeadChanges")
				}
			}()
			return nil
		},
	})
	return nd
}

// In order to resolve the timeout does not work
func handleTimeout(ctx context.Context, f interface{}, args []interface{}) (interface{}, error) {
	if reflect.ValueOf(f).Kind() != reflect.Func {
		return nil, fmt.Errorf("first parameter must be method")
	}

	var out []reflect.Value
	callDone := make(chan struct{})
	rvs := make([]reflect.Value, 0, len(args)+1)
	rvs = append(rvs, reflect.ValueOf(ctx))
	for _, arg := range args {
		rvs = append(rvs, reflect.ValueOf(arg))
	}
	go func() {
		out = reflect.ValueOf(f).Call(rvs)
		close(callDone)
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-callDone:
	}

	if len(out) == 2 {
		if out[1].IsNil() {
			return out[0].Interface(), nil
		}
		return nil, out[1].Interface().(error)
	}

	return nil, fmt.Errorf("method must has 2 return as result")
}

func newMessagePubSubIndirect(logger *log.Logger, networkName types.NetworkName, bootstrap *config.BootstrapConfig, srvCfg config.MessageServiceConfig) (*MessagePubSub, error) {
	return NewMessagePubSub(logger, networkName, bootstrap.Addresses, srvCfg.PublishMessageByPubsub)
}
