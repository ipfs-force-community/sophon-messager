package sqlite

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"
	venustypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/testhelper"
)

func TestAddress(t *testing.T) {
	ctx := context.Background()
	addressRepo := setupRepo(t).AddressRepo()

	rand.Seed(time.Now().Unix())
	addr, err := address.NewIDAddress(rand.Uint64() / 2)
	assert.NoError(t, err)
	addr2, err := address.NewIDAddress(rand.Uint64() / 2)
	assert.NoError(t, err)

	addrInfo := &types.Address{
		ID:                venustypes.NewUUID(),
		Addr:              addr,
		Nonce:             3,
		Weight:            100,
		SelMsgNum:         1,
		State:             types.AddressStateAlive,
		GasOverEstimation: 1.25,
		GasOverPremium:    1.6,
		MaxFee:            big.NewInt(10),
		GasFeeCap:         big.NewInt(1),
		IsDeleted:         -1,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	addrInfo2 := &types.Address{
		ID:             venustypes.NewUUID(),
		Addr:           addr2,
		SelMsgNum:      10,
		State:          types.AddressStateAlive,
		GasOverPremium: 3.0,
		MaxFee:         big.NewInt(110),
		GasFeeCap:      big.NewInt(11),
		Nonce:          2,
		IsDeleted:      -1,
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
	}

	addrInfo3 := &types.Address{
		ID:        venustypes.NewUUID(),
		Addr:      addr,
		Nonce:     3,
		Weight:    1000,
		SelMsgNum: 10,
		State:     types.AddressStateAlive,
		IsDeleted: -1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	addrInfoMap := testhelper.SliceToMap([]*types.Address{addrInfo, addrInfo2})

	t.Run("SaveAddress", func(t *testing.T) {
		assert.NoError(t, addressRepo.SaveAddress(ctx, addrInfo))
		assert.NoError(t, addressRepo.SaveAddress(ctx, addrInfo2))
		assert.Error(t, addressRepo.SaveAddress(ctx, addrInfo3))
	})

	t.Run("ListAddress", func(t *testing.T) {
		addrInfos, err := addressRepo.ListAddress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(addrInfos))

		for _, info := range addrInfos {
			val := addrInfoMap[info.ID.String()]
			testhelper.Equal(t, val, info)
		}
	})

	t.Run("GetAddress", func(t *testing.T) {
		r, err := addressRepo.GetAddress(ctx, addrInfo.Addr)
		assert.NoError(t, err)
		testhelper.Equal(t, addrInfo, r)

		r2, err2 := addressRepo.GetAddress(ctx, address.Undef)
		assert.True(t, errors.Is(err2, gorm.ErrRecordNotFound))
		assert.Nil(t, r2)
	})

	t.Run("GetAddressByID", func(t *testing.T) {
		r, err := addressRepo.GetAddressByID(ctx, addrInfo2.ID)
		assert.NoError(t, err)
		testhelper.Equal(t, addrInfo2, r)
	})

	t.Run("UpdateNonce", func(t *testing.T) {
		nonce := uint64(5)
		assert.NoError(t, addressRepo.UpdateNonce(ctx, addrInfo.Addr, nonce))
		r, err := addressRepo.GetAddress(ctx, addrInfo.Addr)
		assert.NoError(t, err)
		assert.Equal(t, nonce, r.Nonce)

		// set nonce for a not exist address
		err = addressRepo.UpdateNonce(ctx, testhelper.RandAddresses(t, 1)[0], nonce)
		assert.NoError(t, err)
	})

	t.Run("UpdateState", func(t *testing.T) {
		state := types.AddressStateForbbiden
		assert.NoError(t, addressRepo.UpdateState(ctx, addrInfo.Addr, state))
		r, err := addressRepo.GetAddress(ctx, addrInfo.Addr)
		assert.NoError(t, err)
		assert.Equal(t, state, r.State)

		// set state for a not exist address
		err = addressRepo.UpdateState(ctx, testhelper.RandAddresses(t, 1)[0], state)
		assert.NoError(t, err)
	})

	t.Run("UpdateSelectMsgNum", func(t *testing.T) {
		num := uint64(100)
		assert.NoError(t, addressRepo.UpdateSelectMsgNum(ctx, addrInfo.Addr, num))
		r, err := addressRepo.GetAddress(ctx, addrInfo.Addr)
		assert.NoError(t, err)
		assert.Equal(t, num, r.SelMsgNum)

		// set select message count for a not exist address
		err = addressRepo.UpdateSelectMsgNum(ctx, testhelper.RandAddresses(t, 1)[0], num)
		assert.NoError(t, err)
	})

	t.Run("UpdateFeeParams", func(t *testing.T) {
		gasOverEstimation := 1.5
		gasFeeCap := big.NewInt(1000)
		maxFee := big.NewInt(1000)
		gasOverPremium := 1.2
		assert.NoError(t, addressRepo.UpdateFeeParams(ctx, addr, gasOverEstimation, gasOverPremium, maxFee, gasFeeCap))

		r, err := addressRepo.GetAddress(ctx, addr)
		assert.NoError(t, err)
		assert.Equal(t, gasOverEstimation, r.GasOverEstimation)
		assert.Equal(t, maxFee, r.MaxFee)
		assert.Equal(t, gasFeeCap, r.GasFeeCap)
		assert.Equal(t, gasOverPremium, r.GasOverPremium)

		// set fee params for a not exist address
		err = addressRepo.UpdateFeeParams(ctx, testhelper.RandAddresses(t, 1)[0], gasOverEstimation, gasOverPremium, maxFee, gasFeeCap)
		assert.NoError(t, err)
	})

	t.Run("DelAddress", func(t *testing.T) {
		assert.NoError(t, addressRepo.DelAddress(ctx, addrInfo2.Addr))

		r, err := addressRepo.GetAddress(ctx, addrInfo2.Addr)
		assert.Error(t, err)
		assert.Nil(t, r)

		r, err = addressRepo.GetOneRecord(ctx, addrInfo2.Addr)
		assert.NoError(t, err)
		assert.Equal(t, types.AddressStateRemoved, r.State)
		assert.Equal(t, repo.Deleted, r.IsDeleted)

		// delete a not exist address
		err = addressRepo.DelAddress(ctx, testhelper.RandAddresses(t, 1)[0])
		assert.NoError(t, err)
	})

	t.Run("HasAddress", func(t *testing.T) {
		has, err := addressRepo.HasAddress(ctx, addrInfo.Addr)
		assert.NoError(t, err)
		assert.True(t, has)

		has, err = addressRepo.HasAddress(ctx, addrInfo2.Addr)
		assert.NoError(t, err)
		assert.False(t, has)
	})

	t.Run("ListActiveAddress", func(t *testing.T) {
		rs, err := addressRepo.ListActiveAddress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(rs))

		assert.NoError(t, addressRepo.UpdateState(ctx, addrInfo.Addr, types.AddressStateAlive))
		rs, err = addressRepo.ListActiveAddress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(rs))
	})
}
