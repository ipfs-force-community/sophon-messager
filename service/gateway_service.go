package service

import (
	"github.com/filecoin-project/venus-messager/config"
	"github.com/ipfs-force-community/venus-gateway/proofevent"
	"github.com/ipfs-force-community/venus-gateway/walletevent"
)

type GatewayService struct {
	*proofevent.ProofEventStream
	*walletevent.WalletEventStream
}

func NewGatewayService(cfg *config.GatewayConfig) *GatewayService {
	return &GatewayService{
		proofevent.NewProofEventStream(&cfg.Cfg),
		walletevent.NewWalletEventStream(&cfg.Cfg),
	}
}
