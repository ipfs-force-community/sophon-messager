package sqlite

import (
	"context"
	"testing"

	shared "github.com/filecoin-project/venus/venus-shared/types"

	"github.com/filecoin-project/venus/venus-shared/testutil"

	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/stretchr/testify/assert"
)

func Test_fromActorCfg(t *testing.T) {
	ctx := context.Background()
	actorCfgRepo := setupRepo(t).ActorCfgRepo()

	var expectActorCfgs []*types.ActorCfg
	for i := 0; i < 10; i++ {
		var actorCfg types.ActorCfg
		testutil.Provide(t, &actorCfg)
		assert.NoError(t, actorCfgRepo.SaveActorCfg(ctx, &actorCfg))
		expectActorCfgs = append(expectActorCfgs, &actorCfg)

		actorCfgCp := actorCfg
		actorCfgCp.ID = shared.NewUUID()
		err := actorCfgRepo.SaveActorCfg(ctx, &actorCfgCp)
		assert.EqualError(t, err, "UNIQUE constraint failed: actor_cfg.code, actor_cfg.method")

		actorCfg2, err := actorCfgRepo.GetActorCfgByID(ctx, actorCfg.ID)
		assert.NoError(t, err)
		assertActorCfgValue(t, &actorCfg, actorCfg2)
	}

	//ListActorCfg
	actorsList, err := actorCfgRepo.ListActorCfg(ctx)
	assert.NoError(t, err)
	assertActorCfgArrValue(t, expectActorCfgs, actorsList)

	//GetActorCfgByMethodType
	for _, actorCfg := range expectActorCfgs {
		actorActorCfg, err := actorCfgRepo.GetActorCfgByMethodType(ctx, &types.MethodType{
			Code:   actorCfg.Code,
			Method: actorCfg.Method,
		})
		assert.NoError(t, err)
		assertActorCfgValue(t, actorCfg, actorActorCfg)
	}

	//UpdateSelectSpec
	for _, actorCfg := range expectActorCfgs {
		updateAsset := func(cfg func() (*types.ActorCfg, *types.ChangeGasSpecParams)) {
			expectActorCfg, changeParams := cfg()
			err := actorCfgRepo.UpdateSelectSpecById(ctx, actorCfg.ID, changeParams)
			assert.NoError(t, err)

			actorCfg2, err := actorCfgRepo.GetActorCfgByID(ctx, actorCfg.ID)
			assert.NoError(t, err)

			assertActorCfgValue(t, expectActorCfg, actorCfg2)
		}
		var selectSpec types.FeeSpec
		testutil.Provide(t, &selectSpec)
		updateAsset(func() (*types.ActorCfg, *types.ChangeGasSpecParams) {
			actorCfg.GasOverEstimation = selectSpec.GasOverEstimation
			actorCfgCp := *actorCfg
			return &actorCfgCp, &types.ChangeGasSpecParams{
				GasOverEstimation: &selectSpec.GasOverEstimation,
			}
		})

		updateAsset(func() (*types.ActorCfg, *types.ChangeGasSpecParams) {
			actorCfg.GasOverPremium = selectSpec.GasOverPremium
			actorCfgCp := *actorCfg
			return &actorCfgCp, &types.ChangeGasSpecParams{
				GasOverPremium: &selectSpec.GasOverPremium,
			}
		})

		updateAsset(func() (*types.ActorCfg, *types.ChangeGasSpecParams) {
			actorCfg.MaxFee = selectSpec.MaxFee
			actorCfgCp := *actorCfg
			return &actorCfgCp, &types.ChangeGasSpecParams{
				MaxFee: selectSpec.MaxFee,
			}
		})

		updateAsset(func() (*types.ActorCfg, *types.ChangeGasSpecParams) {
			actorCfg.GasFeeCap = selectSpec.GasFeeCap
			actorCfgCp := *actorCfg
			return &actorCfgCp, &types.ChangeGasSpecParams{
				GasFeeCap: selectSpec.GasFeeCap,
			}
		})
		updateAsset(func() (*types.ActorCfg, *types.ChangeGasSpecParams) {
			actorCfg.BaseFee = selectSpec.BaseFee
			actorCfgCp := *actorCfg
			return &actorCfgCp, &types.ChangeGasSpecParams{
				BaseFee: selectSpec.BaseFee,
			}
		})
	}

	//Delete

	for _, actorCfg := range expectActorCfgs[:5] {
		assert.NoError(t, actorCfgRepo.DelActorCfgById(ctx, actorCfg.ID))
	}

	for _, actorCfg := range expectActorCfgs[5:] {
		assert.NoError(t, actorCfgRepo.DelActorCfgByMethodType(ctx, &types.MethodType{actorCfg.Code, actorCfg.Method}))
	}

	actorsR, err := actorCfgRepo.ListActorCfg(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(actorsR))
}

func assertActorCfgValue(t *testing.T, expectVal, actualVal *types.ActorCfg) {
	assert.Equal(t, expectVal.ID, actualVal.ID)
	assert.Equal(t, expectVal.ActorVersion, actualVal.ActorVersion)
	assert.Equal(t, expectVal.MethodType, actualVal.MethodType)
	assert.Equal(t, expectVal.FeeSpec, actualVal.FeeSpec)
}

func assertActorCfgArrValue(t *testing.T, expectVal, actualVal []*types.ActorCfg) {
	assert.Equal(t, len(expectVal), len(actualVal))

	for index, val := range expectVal {
		assertActorCfgValue(t, val, actualVal[index])
	}
}
