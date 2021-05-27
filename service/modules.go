package service

import (
	"context"
	"reflect"
	"time"

	"golang.org/x/xerrors"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type ServiceMap map[reflect.Type]interface{}

func MakeServiceMap(msgService *MessageService,
	addressService *AddressService,
	sps *SharedParamsService,
	nodeService *NodeService) ServiceMap {
	sMap := make(ServiceMap)
	sMap[reflect.TypeOf(msgService)] = msgService
	sMap[reflect.TypeOf(addressService)] = addressService
	sMap[reflect.TypeOf(sps)] = sps
	sMap[reflect.TypeOf(nodeService)] = nodeService
	return sMap
}

func MessagerService() fx.Option {
	return fx.Options(
		fx.Provide(NewMessageService),
		//fx.Provide(NewWalletService),
		fx.Provide(NewAddressService),
		fx.Provide(NewSharedParamsService),
		fx.Provide(NewNodeService),
		fx.Provide(MakeServiceMap),
	)
}

func StartNodeEvents(lc fx.Lifecycle, client *NodeClient, msgService *MessageService, log *logrus.Logger) *NodeEvents {
	nd := &NodeEvents{
		client:     client,
		log:        log,
		msgService: msgService,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if !msgService.cfg.SkipPushMessage {
				go msgService.StartPushMessage(ctx)
			} else {
				msgService.log.Infof("skip push message")
			}
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
func handleTimeout(f interface{}, ctx context.Context, args []interface{}) (interface{}, error) {
	if reflect.ValueOf(f).Kind() != reflect.Func {
		return nil, xerrors.Errorf("first parameter must be method")
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

	return nil, xerrors.Errorf("method must has 2 return as result")
}
