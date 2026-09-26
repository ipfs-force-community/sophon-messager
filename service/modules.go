package service

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"

	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
)

var log = logging.Logger("service")

func MessagerService() fx.Option {
	return fx.Options(
		fx.Provide(NewMessageService),
		fx.Provide(NewAddressService),
		fx.Provide(NewSharedParamsService),
		fx.Provide(NewINodeService),
		fx.Provide(NewNodeService),
	)
}

func StartNodeEvents(lc fx.Lifecycle, client v1.FullNode, msgService *MessageService) *NodeEvents {
	nd := &NodeEvents{
		client:     client,
		msgService: msgService,
	}

	// These loops outlive the start hook, so they must not run on the hook's
	// context: fx cancels that one only after every stop hook has returned, and a
	// caller may pass a context that is never cancelled at all.
	ctx, cancel := context.WithCancel(context.Background())

	var nodeEventsWg sync.WaitGroup
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			nodeEventsWg.Add(2)
			go func() {
				defer nodeEventsWg.Done()
				msgService.StartPushMessage(ctx, msgService.fsRepo.Config().MessageService.SkipPushMessage)
			}()
			go func() {
				defer nodeEventsWg.Done()
				for {
					if err := nd.listenHeadChangesOnce(ctx); err != nil {
						log.Errorf("listen head changes errored: %s", err)
					} else {
						log.Warn("listenHeadChanges quit")
					}
					select {
					case <-time.After(time.Second):
					case <-ctx.Done():
						log.Warnf("stop listen head changes: %s", ctx.Err())
						return
					}

					log.Info("restarting listenHeadChanges")
				}
			}()
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			cancel()
			joinGroup(stopCtx, "node events", &nodeEventsWg)
			return msgService.Close(stopCtx)
		},
	})
	return nd
}

// shutdownTimeout bounds every wait for a background loop, so that a loop
// ignoring cancellation cannot block app.Stop.
var shutdownTimeout = 10 * time.Second

// joinGroup waits for wg to drain, giving up after shutdownTimeout. Repo writers
// must be joined before their data directory is removed: a write landing inside
// the removal leaves the directory non-empty. A loop that ignores cancellation
// must not fail shutdown, so giving up is logged and shutdown continues.
func joinGroup(ctx context.Context, name string, wg *sync.WaitGroup) {
	ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		log.Warnf("%s: background loops still running after %s", name, shutdownTimeout)
	}
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
