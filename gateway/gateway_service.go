package gateway

import (
	"context"

	"github.com/ipfs-force-community/venus-gateway/walletevent"

	"github.com/filecoin-project/venus-messager/config"
)

type GatewayService struct {
	*walletevent.WalletEventStream
}

var _ IWalletClient = &GatewayService{}

func NewGatewayService(cfg *config.GatewayConfig) *GatewayService {
	return &GatewayService{
		walletevent.NewWalletEventStream(context.Background(), nil, &cfg.Cfg),
	}
}
