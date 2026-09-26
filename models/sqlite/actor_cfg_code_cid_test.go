package sqlite

import (
	"context"
	"testing"

	"github.com/filecoin-project/go-state-types/actors"
	shared "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
)

// Mainnet account actor code CIDs of actor version 18 and 19.
const (
	accountCodeV18 = "bafk2bzacebt5o43hdveql4tqrnr7bg62by3d336ddsfl3l72xedhlsrbe6nia"
	accountCodeV19 = "bafk2bzacecgttsyulmd4r2y64iw64x2t4w7tkh636tbtkemy37ume6fxtcjuu"
)

// A config stored while the network ran actor version 18 keeps matching only
// that code CID, so the version 19 code CID the chain reports after the upgrade
// misses without an error and the gas parameters stop applying.
func TestActorCfgLookupNeedsRowPerCodeCid(t *testing.T) {
	ctx := context.Background()
	actorCfgRepo := setupRepo(t).ActorCfgRepo()

	codeV18, err := cid.Decode(accountCodeV18)
	require.NoError(t, err)
	codeV19, err := cid.Decode(accountCodeV19)
	require.NoError(t, err)

	require.NoError(t, actorCfgRepo.SaveActorCfg(ctx, &types.ActorCfg{
		ID:           shared.NewUUID(),
		ActorVersion: actors.Version18,
		MethodType:   types.MethodType{Code: codeV18, Method: 0},
		FeeSpec:      types.FeeSpec{GasOverEstimation: 1.25},
	}))

	has, err := actorCfgRepo.HasActorCfg(ctx, &types.MethodType{Code: codeV19, Method: 0})
	require.NoError(t, err)
	require.False(t, has)

	require.NoError(t, actorCfgRepo.SaveActorCfg(ctx, &types.ActorCfg{
		ID:           shared.NewUUID(),
		ActorVersion: actors.Version19,
		MethodType:   types.MethodType{Code: codeV19, Method: 0},
		FeeSpec:      types.FeeSpec{GasOverEstimation: 1.25},
	}))

	has, err = actorCfgRepo.HasActorCfg(ctx, &types.MethodType{Code: codeV19, Method: 0})
	require.NoError(t, err)
	require.True(t, has)

	cfg, err := actorCfgRepo.GetActorCfgByMethodType(ctx, &types.MethodType{Code: codeV19, Method: 0})
	require.NoError(t, err)
	require.Equal(t, 1.25, cfg.GasOverEstimation)
	require.Equal(t, actors.Version19, cfg.ActorVersion)
}
