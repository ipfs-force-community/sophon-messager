package utils

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/builtin/v10/eam"
	"github.com/filecoin-project/go-state-types/proof"
	miner5 "github.com/filecoin-project/specs-actors/v5/actors/builtin/miner"
	"github.com/filecoin-project/venus/venus-shared/testutil"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/stretchr/testify/assert"
)

func TestMerge(t *testing.T) {
	data := []uint64{0, 1, 4, 5, 6, 9, 12, 16, 17}
	expect := [][]uint64{{0, 1}, {4, 6}, {9}, {12}, {16, 17}}
	assert.Equal(t, expect, merge(data))

	data2 := []uint64{0, 4, 6, 7, 12, 16}
	expect2 := [][]uint64{{0}, {4}, {6, 7}, {12}, {16}}
	assert.Equal(t, expect2, merge(data2))
}

func TestConvertBitFieldToString(t *testing.T) {
	data := bitfield.NewFromSet([]uint64{1, 4, 5, 9, 15, 16})
	res, err := convertBitFieldToString(data)
	assert.NoError(t, err)
	except := "1, 4-5, 9, 15-16"
	assert.Equal(t, except, res)

	data2 := bitfield.NewFromSet([]uint64{1, 4, 9, 15})
	res2, err2 := convertBitFieldToString(data2)
	assert.NoError(t, err2)
	except2 := "1, 4, 9, 15"
	assert.Equal(t, except2, res2)
}

func TestTryConvertParams(t *testing.T) {
	t.Run("test convert ExtendSectorExpirationParams", func(t *testing.T) {
		params := &types.ExtendSectorExpirationParams{
			Extensions: []types.ExpirationExtension{
				{
					Deadline:      10,
					Partition:     101,
					Sectors:       bitfield.NewFromSet([]uint64{1, 2, 3, 4}),
					NewExpiration: 100,
				},
			},
		}

		res, err := TryConvertParams(params)
		assert.NoError(t, err)
		val, ok := res.(map[string]interface{})
		assert.True(t, ok)
		expect := []map[string]interface{}{
			{
				"Deadline":      10,
				"Partition":     101,
				"Sectors":       "1-4",
				"NewExpiration": 100,
			},
		}
		equalMarshal(t, expect, val["Extensions"])
	})

	t.Run("test convert DeclareFaultsRecoveredParams", func(t *testing.T) {
		params := &types.DeclareFaultsRecoveredParams{
			Recoveries: []types.RecoveryDeclaration{
				{
					Deadline:  10,
					Partition: 100,
					Sectors:   bitfield.NewFromSet([]uint64{1, 4, 7}),
				},
			},
		}

		res, err := TryConvertParams(params)
		assert.NoError(t, err)
		val, ok := res.(map[string]interface{})
		assert.True(t, ok)
		expect := []map[string]interface{}{
			{
				"Deadline":  10,
				"Partition": 100,
				"Sectors":   "1, 4, 7",
			},
		}
		equalMarshal(t, expect, val["Recoveries"])
	})

	t.Run("test convert DeclareFaultsParams", func(t *testing.T) {
		params := &types.DeclareFaultsParams{
			Faults: []types.FaultDeclaration{
				{
					Deadline:  10,
					Partition: 100,
					Sectors:   bitfield.NewFromSet([]uint64{1, 4, 7}),
				},
			},
		}

		res, err := TryConvertParams(params)
		assert.NoError(t, err)
		val, ok := res.(map[string]interface{})
		assert.True(t, ok)
		expect := []map[string]interface{}{
			{
				"Deadline":  10,
				"Partition": 100,
				"Sectors":   "1, 4, 7",
			},
		}
		equalMarshal(t, expect, val["Faults"])
	})

	t.Run("test convert ProveCommitAggregateParams", func(t *testing.T) {
		params := &types.ProveCommitAggregateParams{
			SectorNumbers:  bitfield.NewFromSet([]uint64{1, 7}),
			AggregateProof: []byte("AggregateProof"),
		}

		res, err := TryConvertParams(params)
		assert.NoError(t, err)
		expect := map[string]interface{}{
			"SectorNumbers":  "1, 7",
			"AggregateProof": []byte("AggregateProof"),
		}
		equalMarshal(t, expect, res)
	})

	t.Run("test convert TerminateSectorsParams", func(t *testing.T) {
		params := &types.TerminateSectorsParams{
			Terminations: []types.TerminationDeclaration{
				{
					Deadline:  10,
					Partition: 100,
					Sectors:   bitfield.NewFromSet([]uint64{1, 4}),
				},
			},
		}

		res, err := TryConvertParams(params)
		assert.NoError(t, err)
		val, ok := res.(map[string]interface{})
		assert.True(t, ok)
		expect := []map[string]interface{}{
			{
				"Deadline":  10,
				"Partition": 100,
				"Sectors":   "1, 4",
			},
		}
		equalMarshal(t, expect, val["Terminations"])
	})

	t.Run("test convert CompactPartitionsParams", func(t *testing.T) {
		params := &types.CompactPartitionsParams{
			Deadline:   100,
			Partitions: bitfield.NewFromSet([]uint64{1, 2, 8, 10}),
		}

		res, err := TryConvertParams(params)
		assert.NoError(t, err)
		expect := map[string]interface{}{
			"Deadline":   100,
			"Partitions": "1-2, 8, 10",
		}
		equalMarshal(t, expect, res)
	})

	t.Run("test convert CompactSectorNumbersParams", func(t *testing.T) {
		params := &types.CompactSectorNumbersParams{
			MaskSectorNumbers: bitfield.NewFromSet([]uint64{1, 2, 8, 10}),
		}

		res, err := TryConvertParams(params)
		assert.NoError(t, err)
		expect := map[string]interface{}{
			"MaskSectorNumbers": "1-2, 8, 10",
		}
		equalMarshal(t, expect, res)
	})

	t.Run("test convert SubmitWindowedPoStParams", func(t *testing.T) {
		params := &types.SubmitWindowedPoStParams{
			Deadline: 10,
			Partitions: []types.PoStPartition{
				{
					Index:   101,
					Skipped: bitfield.NewFromSet([]uint64{2, 3, 6, 7}),
				},
			},
			ChainCommitEpoch: 100,
			ChainCommitRand:  []byte("ChainCommitRand"),
		}

		res, err := TryConvertParams(params)
		assert.NoError(t, err)
		expect := map[string]interface{}{
			"Deadline": 10,
			"Partitions": []map[string]interface{}{
				{
					"Index":   101,
					"Skipped": "2-3, 6-7",
				},
			},
			"Proofs":           []proof.PoStProof{},
			"ChainCommitEpoch": 100,
			"ChainCommitRand":  []byte("ChainCommitRand"),
		}
		equalMarshal(t, expect, res)
	})

	t.Run("test convert PublishStorageDealsReturn", func(t *testing.T) {
		params := &types.PublishStorageDealsReturn{
			IDs:        []abi.DealID{1, 5},
			ValidDeals: bitfield.NewFromSet([]uint64{1, 2, 8, 10}),
		}

		res, err := TryConvertParams(params)
		assert.NoError(t, err)
		expect := map[string]interface{}{
			"IDs":        []abi.DealID{1, 5},
			"ValidDeals": "1-2, 8, 10",
		}
		equalMarshal(t, expect, res)
	})

	t.Run("test convert CreateReturn", func(t *testing.T) {
		addr := testutil.AddressProvider()(t)
		ethAddr := [20]byte{92, 41, 100, 227, 56, 253, 50, 107, 148, 193, 190, 112, 83, 185, 97, 122, 70, 160, 60, 177}
		cr := eam.CreateReturn{
			ActorID:       5917,
			RobustAddress: &addr,
			EthAddress:    ethAddr,
		}
		res, err := TryConvertParams(cr)
		assert.NoError(t, err)
		expect := map[string]interface{}{
			"ActorID":       5917,
			"RobustAddress": addr.String(),
			"EthAddress":    "0x5c2964e338fd326b94c1be7053b9617a46a03cb1",
		}
		equalMarshal(t, expect, res)

		cr2 := eam.CreateReturn{
			ActorID:       0,
			RobustAddress: nil,
			EthAddress:    [20]byte{},
		}
		res2, err := TryConvertParams(cr2)
		assert.NoError(t, err)
		expect2 := map[string]interface{}{
			"ActorID":       0,
			"RobustAddress": nil,
			"EthAddress":    "0x0000000000000000000000000000000000000000",
		}
		equalMarshal(t, expect2, res2)
	})

	t.Run("test convert ActiveBeneficiary", func(t *testing.T) {
		addr := testutil.AddressProvider()(t)
		ab := &types.ActiveBeneficiary{
			Beneficiary: addr,
			Term: types.BeneficiaryTerm{
				Quota:      abi.NewTokenAmount(100),
				UsedQuota:  abi.NewTokenAmount(100),
				Expiration: 100,
			},
		}
		res, err := TryConvertParams(ab)
		assert.NoError(t, err)
		assert.Equal(t, ab, res)
	})
}

func equalMarshal(t *testing.T, expect, actual interface{}) {
	d, err := json.Marshal(expect)
	assert.NoError(t, err)
	d2, err := json.Marshal(actual)
	assert.NoError(t, err)
	assert.Equal(t, string(d), string(d2))
}

func TestHasBitfield(t *testing.T) {
	cases := []struct {
		typ    any
		expect bool
	}{
		{&types.ExtendSectorExpirationParams{}, true},
		{&miner5.ExtendSectorExpirationParams{}, true},
		{&types.DeclareFaultsRecoveredParams{}, true},
		{&miner5.DeclareFaultsRecoveredParams{}, true},
		{&types.DeclareFaultsParams{}, true},
		{&miner5.ProveCommitAggregateParams{}, true},
		{&types.TerminateSectorsParams{}, true},
		{&types.CompactPartitionsParams{}, true},
		{&types.CompactSectorNumbersParams{}, true},
		{&types.SubmitWindowedPoStParams{}, true},
		{&types.PublishStorageDealsReturn{}, true},
		{&types.ActiveBeneficiary{}, false},
		{&types.ActivateDealsParams{}, false},
		{&eam.CreateReturn{}, true},
	}
	for _, c := range cases {
		testutil.Provide(t, c.typ, testutil.BytesFixedProvider(20))
		if actual := isNeedConvert(reflect.ValueOf(c.typ)); actual != c.expect {
			t.Errorf("call %T failed, actual %v, expect %v", c.typ, actual, c.expect)
		}
	}
}
