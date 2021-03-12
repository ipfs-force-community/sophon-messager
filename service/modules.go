package service

import (
	"context"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type ServiceMap map[reflect.Type]interface{}

func MakeServiceMap(msgService *MessageService, walletService *WalletService, addressService *AddressService) ServiceMap {
	sMap := make(ServiceMap)
	sMap[reflect.TypeOf(msgService)] = msgService
	sMap[reflect.TypeOf(walletService)] = walletService
	sMap[reflect.TypeOf(addressService)] = addressService
	return sMap
}

func MessagerService() fx.Option {
	return fx.Options(
		fx.Provide(NewMessageService),
		fx.Provide(NewWalletService),
		fx.Provide(NewAddressService),
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
			go msgService.StartPushMessage(ctx)
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
