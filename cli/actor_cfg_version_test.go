package cli

import (
	"testing"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/actors"
	"github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
)

// Builtin actor code CIDs change with the actor version, so a config row keyed
// by the v18 code CID stops matching once nv29 activates the v19 code CID.
const (
	mainnetAccountV18     = "bafk2bzacebt5o43hdveql4tqrnr7bg62by3d336ddsfl3l72xedhlsrbe6nia"
	mainnetAccountV19     = "bafk2bzacecgttsyulmd4r2y64iw64x2t4w7tkh636tbtkemy37ume6fxtcjuu"
	calibrationAccountV18 = "bafk2bzaceaaekrmq3cqnzokwmuksgr7vqgslpm67blh3fhj4cwgnk5vgoksam"
)

func mustCode(t *testing.T, code string) cid.Cid {
	t.Helper()

	c, err := cid.Decode(code)
	require.NoError(t, err)

	return c
}

func actorCfgFixture(code cid.Cid, method abi.MethodNum) *messager.ActorCfg {
	return &messager.ActorCfg{
		ActorVersion: actors.Version18,
		MethodType:   messager.MethodType{Code: code, Method: method},
		FeeSpec:      messager.FeeSpec{GasOverEstimation: 1.25},
	}
}

func TestNewestActorVersion(t *testing.T) {
	version, err := newestActorVersion("mainnet")
	require.NoError(t, err)
	require.Equal(t, actors.Version19, version)

	_, err = newestActorVersion("no-such-network")
	require.Error(t, err)
}

func TestAnalyzeActorCfgVersionFlagsRowPinnedToSupersededCode(t *testing.T) {
	cfgs := []*messager.ActorCfg{
		actorCfgFixture(mustCode(t, mainnetAccountV18), 0),
		actorCfgFixture(mustCode(t, mainnetAccountV19), 1),
	}

	status, err := analyzeActorCfgVersion(cfgs, "mainnet", actors.Version19)
	require.NoError(t, err)

	require.Len(t, status.Superseded, 1)
	gap := status.Superseded[0]
	require.Equal(t, "account", gap.ActorName)
	require.Equal(t, abi.MethodNum(0), gap.Method)
	require.True(t, mustCode(t, mainnetAccountV19).Equals(gap.TargetCode))
	require.False(t, gap.HasTargetRow)
	require.Equal(t, 1, status.Current)
	require.Len(t, status.pending(), 1)
}

func TestAnalyzeActorCfgVersionClearsRowThatHasTargetRow(t *testing.T) {
	cfgs := []*messager.ActorCfg{
		actorCfgFixture(mustCode(t, mainnetAccountV18), 0),
		actorCfgFixture(mustCode(t, mainnetAccountV19), 0),
	}

	status, err := analyzeActorCfgVersion(cfgs, "mainnet", actors.Version19)
	require.NoError(t, err)

	require.Len(t, status.Superseded, 1)
	require.True(t, status.Superseded[0].HasTargetRow)
	require.Empty(t, status.pending())
}

func TestAnalyzeActorCfgVersionReportsCodeOfAnotherNetwork(t *testing.T) {
	cfgs := []*messager.ActorCfg{actorCfgFixture(mustCode(t, calibrationAccountV18), 0)}

	status, err := analyzeActorCfgVersion(cfgs, "mainnet", actors.Version19)
	require.NoError(t, err)

	require.Len(t, status.Foreign, 1)
	require.Empty(t, status.Superseded)
	require.Zero(t, status.Current)
}

func TestAnalyzeActorCfgVersionRejectsUnknownActorVersion(t *testing.T) {
	_, err := analyzeActorCfgVersion(nil, "mainnet", actors.Version(7))
	require.Error(t, err)
}

// Fifteen of the sixteen mainnet builtin actors change code CID between actor
// version 18 and 19, so a config set written under version 18 is superseded
// almost entirely by the upgrade.
func TestAnalyzeActorCfgVersionFlagsChangedActorsOfMainnet(t *testing.T) {
	var cfgs []*messager.ActorCfg
	for name, code := range builtinActorCodesByVersion("mainnet")[actors.Version18] {
		require.NotEmpty(t, name)
		cfgs = append(cfgs, actorCfgFixture(code, 0))
	}

	status, err := analyzeActorCfgVersion(cfgs, "mainnet", actors.Version19)
	require.NoError(t, err)

	require.Len(t, status.Superseded, 15)
	require.Equal(t, 1, status.Current)
	require.Len(t, status.pending(), 15)
	require.Empty(t, status.Foreign)
	require.Empty(t, status.NoTarget)
}
