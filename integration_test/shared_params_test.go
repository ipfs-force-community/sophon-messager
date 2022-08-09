package integration

import (
	"context"
	"testing"
	"time"

	"github.com/filecoin-project/go-state-types/big"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/service"
)

func TestSharedParamsAPI(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultConfig()
	cfg.API.Address = "/ip4/0.0.0.0/tcp/0"
	cfg.MessageService.SkipPushMessage = true
	cfg.MessageService.WaitingChainHeadStableDuration = 2 * time.Second
	ms, err := mockMessagerServer(ctx, t.TempDir(), cfg)
	assert.NoError(t, err)

	go ms.start(ctx)
	assert.NoError(t, <-ms.appStartErr)

	cli, closer, err := newMessagerClient(ctx, ms.port, ms.token)
	assert.NoError(t, err)
	defer closer()

	res, err := cli.GetSharedParams(ctx)
	assert.NoError(t, err)
	assert.Equal(t, service.DefSharedParams, res)

	params := &types.SharedSpec{
		ID:                1,
		GasOverEstimation: 10,
		MaxFee:            big.NewInt(11111111),
		GasFeeCap:         big.NewInt(11111112),
		GasOverPremium:    10,
		SelMsgNum:         100,
	}
	assert.NoError(t, cli.SetSharedParams(ctx, params))

	res, err = cli.GetSharedParams(ctx)
	assert.NoError(t, err)
	assert.Equal(t, params, res)

	assert.NoError(t, ms.stop(ctx))
}
