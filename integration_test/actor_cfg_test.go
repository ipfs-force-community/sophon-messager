package integration

import (
	"context"
	"testing"
	"time"

	types2 "github.com/filecoin-project/venus/venus-shared/types"

	"github.com/filecoin-project/venus/venus-shared/testutil"

	"github.com/filecoin-project/venus/venus-shared/api/messager"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/testhelper"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/stretchr/testify/assert"
)

func TestActorCfgAPI(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultConfig()
	cfg.API.Address = "/ip4/0.0.0.0/tcp/0"
	cfg.MessageService.SkipPushMessage = true
	cfg.MessageService.WaitingChainHeadStableDuration = 2 * time.Second
	authClient := testhelper.NewMockAuthClient()
	ms, err := mockMessagerServer(ctx, t.TempDir(), cfg, authClient)
	assert.NoError(t, err)

	go ms.start(ctx)
	assert.NoError(t, <-ms.appStartErr)

	account := defaultLocalToken
	addrCount := 10
	addrs := testhelper.RandAddresses(t, addrCount)
	authClient.AddMockUserAndSigner(account, addrs)
	assert.NoError(t, ms.walletCli.AddAddress(account, addrs))

	api, closer, err := newMessagerClient(ctx, ms.port, ms.token)
	assert.NoError(t, err)
	defer closer()

	actorCfgs := make([]*types.ActorCfg, 5)
	testutil.Provide(t, &actorCfgs)

	t.Run("test save actor config", func(t *testing.T) {
		testCreateActorCfg(ctx, t, api, actorCfgs)
	})

	t.Run("test list actor config", func(t *testing.T) {
		testListActorCfg(ctx, t, api, actorCfgs)
	})

	t.Run("test get actor config", func(t *testing.T) {
		testGetActorCfg(ctx, t, api, actorCfgs)
	})

	t.Run("test update actor config", func(t *testing.T) {
		testUpdateActorCfg(ctx, t, api, actorCfgs)
	})

	assert.NoError(t, ms.stop(ctx))
}

func testCreateActorCfg(ctx context.Context, t *testing.T, api messager.IMessager, actorCfgs []*types.ActorCfg) {
	for _, actorCfg := range actorCfgs {
		assert.NoError(t, api.SaveActorCfg(ctx, actorCfg))
	}
}

func testListActorCfg(ctx context.Context, t *testing.T, api messager.IMessager, actorCfgs []*types.ActorCfg) {
	listResult, err := api.ListActorCfg(ctx)
	assert.NoError(t, err)
	assert.Len(t, listResult, len(actorCfgs))

	expect := map[types2.UUID]*types.ActorCfg{}
	for _, actorCfg := range actorCfgs {
		expect[actorCfg.ID] = actorCfg
	}

	for _, r := range listResult {
		testhelper.Equal(t, expect[r.ID], r)
	}
}

func testGetActorCfg(ctx context.Context, t *testing.T, api messager.IMessager, actorCfgs []*types.ActorCfg) {
	expect := map[types2.UUID]*types.ActorCfg{}
	for _, actorCfg := range actorCfgs {
		expect[actorCfg.ID] = actorCfg
	}

	for _, actorCfg := range actorCfgs {
		getResult, err := api.GetActorCfgByID(ctx, actorCfg.ID)
		assert.NoError(t, err)
		testhelper.Equal(t, expect[getResult.ID], getResult)
	}
}

func testUpdateActorCfg(ctx context.Context, t *testing.T, api messager.IMessager, actorCfgs []*types.ActorCfg) {
	expect := map[types2.UUID]*types.ActorCfg{}
	for _, actorCfg := range actorCfgs {
		expect[actorCfg.ID] = actorCfg
	}

	for _, actorCfg := range actorCfgs {
		changeParams := &types.ChangeGasSpecParams{}
		testutil.Provide(t, changeParams)
		err := api.UpdateActorCfg(ctx, actorCfg.ID, changeParams)
		assert.NoError(t, err)

		updatedR, err := api.GetActorCfgByID(ctx, actorCfg.ID)
		assert.NoError(t, err)
		assert.Equal(t, updatedR.FeeSpec, types.FeeSpec{
			GasOverEstimation: *changeParams.GasOverEstimation,
			MaxFee:            changeParams.MaxFee,
			GasFeeCap:         changeParams.GasFeeCap,
			GasOverPremium:    *changeParams.GasOverPremium,
			BaseFee:           changeParams.BaseFee,
		})
	}
}
