package gateway

import (
	"context"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/ipfs-force-community/venus-gateway/walletevent"
)

type GatewayService struct {
	*walletevent.WalletEventStream
}

func NewGatewayService(cfg *config.GatewayConfig) *GatewayService {
	return &GatewayService{
		walletevent.NewWalletEventStream(context.Background(), &cfg.Cfg),
	}
}
