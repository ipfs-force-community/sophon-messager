package integration

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/ipfs-force-community/sophon-messager/config"
	"github.com/ipfs-force-community/sophon-messager/testhelper"
)

const defaultLocalToken = "defaultLocalToken"

func TestAddressAPI(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultConfig()
	cfg.API.Address = "/ip4/0.0.0.0/tcp/0"
	cfg.MessageService.SkipPushMessage = true
	cfg.MessageService.WaitingChainHeadStableDuration = 2 * time.Second
	authClient := testhelper.NewMockAuthClient(t)
	ms, err := mockMessagerServer(ctx, t.TempDir(), cfg, authClient)
	assert.NoError(t, err)

	go ms.start(ctx)
	assert.NoError(t, <-ms.appStartErr)

	// account with local token of admin perm and 10 signers
	accountLocal := defaultLocalToken
	addrCount := 10
	addrs := testhelper.RandAddresses(t, addrCount)
	authClient.Init(accountLocal, addrs)
	assert.NoError(t, ms.walletCli.AddAddress(accountLocal, addrs))

	apiWithLocalToken, closer, err := newMessagerClient(ctx, ms.port, ms.token)
	assert.NoError(t, err)
	defer closer()

	// account with token of sign permission and no signers
	accountSign := "accountSign"
	playLoad := &auth.JWTPayload{
		Name: accountSign,
		Perm: core.PermSign,
	}
	tokenSign, err := genToken(playLoad)
	assert.NoError(t, err)
	authClient.EXPECT().Verify(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, token string) (*auth.VerifyResponse, error) {
		if token == tokenSign {
			return playLoad, nil
		}
		return nil, errors.New("not exist")
	}).AnyTimes()
	apiSign, closer2, err := newMessagerClient(ctx, ms.port, tokenSign)
	assert.NoError(t, err)
	defer closer2()

	allAddrs := make([]address.Address, 0, len(addrs))
	for _, addr := range addrs {
		allAddrs = append(allAddrs, testhelper.ResolveAddr(t, addr))
	}

	usedAddrs := make(map[address.Address]struct{})
	msgs := testhelper.NewMessages(addrCount * 2)
	addrMsgs := make(map[address.Address][]*types.Message, len(addrs))
	for _, msg := range msgs {
		msg.From = addrs[rand.Intn(addrCount)]
		msg.To = addrs[rand.Intn(addrCount)]
		id, err := apiWithLocalToken.PushMessageWithId(ctx, msg.ID, &msg.Message, msg.Meta)
		assert.NoError(t, err)
		assert.Equal(t, msg.ID, id)

		from := testhelper.ResolveAddr(t, msg.From)
		usedAddrs[from] = struct{}{}
		addrMsgs[from] = append(addrMsgs[from], msg)
	}

	t.Run("test get address and has address", func(t *testing.T) {
		testGetAddressAndHasAddress(ctx, t, apiWithLocalToken, apiSign, allAddrs, usedAddrs)
	})
	t.Run("test wallet has", func(t *testing.T) {
		testWalletHas(ctx, t, apiWithLocalToken, apiSign, allAddrs)
	})
	t.Run("test list address", func(t *testing.T) {
		testListAddress(ctx, t, apiWithLocalToken, apiSign, usedAddrs)
	})
	t.Run("test update nonce", func(t *testing.T) {
		testUpdateNonce(ctx, t, apiWithLocalToken, apiSign, allAddrs)
	})
	t.Run("test forbidden and active address", func(t *testing.T) {
		testForbiddenAndActiveAddress(ctx, t, apiWithLocalToken, apiSign, allAddrs, usedAddrs)
	})
	t.Run("test set select message num", func(t *testing.T) {
		testSetSelectMsgNum(ctx, t, apiWithLocalToken, apiSign, allAddrs, usedAddrs)
	})
	t.Run("test set fee params", func(t *testing.T) {
		testSetFeeParams(ctx, t, apiWithLocalToken, apiSign, allAddrs, usedAddrs)
	})
	t.Run("test clear unfill message", func(t *testing.T) {
		testClearUnFillMessage(ctx, t, apiWithLocalToken, apiSign, allAddrs, addrMsgs)
	})
	t.Run("test delete address", func(t *testing.T) {
		testDeleteAddress(ctx, t, apiWithLocalToken, apiSign, allAddrs, usedAddrs)
	})

	assert.NoError(t, ms.stop(ctx))
}

func testGetAddressAndHasAddress(ctx context.Context,
	t *testing.T,
	api, apiNoPerm messager.IMessager,
	allAddrs []address.Address,
	usedAddrs map[address.Address]struct{},
) {
	var err error
	for _, addr := range allAddrs {
		_, ok := usedAddrs[addr]
		addrInfo, getAddrErr := api.GetAddress(ctx, addr)
		assert.NoError(t, err)
		_, err := apiNoPerm.GetAddress(ctx, addr)
		assert.Contains(t, err.Error(), "permission deny")

		// test has address
		has, err := api.HasAddress(ctx, addr)
		assert.NoError(t, err)
		_, err = apiNoPerm.HasAddress(ctx, addr)
		assert.Contains(t, err.Error(), "permission deny")

		if ok {
			assert.NoError(t, getAddrErr)
			assert.Equal(t, addr, addrInfo.Addr)
			assert.Equal(t, uint64(0), addrInfo.Nonce)
			assert.Equal(t, types.AddressStateAlive, addrInfo.State)
			assert.Equal(t, uint64(0), addrInfo.SelMsgNum)
			assert.Equal(t, 0.0, addrInfo.GasOverEstimation)
			assert.Equal(t, big.Zero(), addrInfo.MaxFee)
			assert.Equal(t, big.Zero(), addrInfo.GasFeeCap)
			assert.Equal(t, 0.0, addrInfo.GasOverPremium)
			assert.Equal(t, big.Zero(), addrInfo.BaseFee)
			assert.True(t, has)
		} else {
			assert.Contains(t, getAddrErr.Error(), gorm.ErrRecordNotFound.Error())
			assert.False(t, has)
		}
	}
}

func testWalletHas(ctx context.Context, t *testing.T, api, apiNoPerm messager.IMessager, allAddrs []address.Address) {
	for _, addr := range allAddrs {
		has, err := api.WalletHas(ctx, addr)
		assert.NoError(t, err)
		assert.True(t, has)

		_, err = apiNoPerm.WalletHas(ctx, addr)
		assert.Contains(t, err.Error(), "permission deny")
	}
}

func testListAddress(ctx context.Context, t *testing.T, api, apiNoPerm messager.IMessager, usedAddrs map[address.Address]struct{}) {
	addrInfos, err := api.ListAddress(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(usedAddrs), len(addrInfos))
	for _, addrInfo := range addrInfos {
		_, ok := usedAddrs[addrInfo.Addr]
		assert.True(t, ok)
		assert.Equal(t, types.AddressStateAlive, addrInfo.State)
	}

	resNUll, err := apiNoPerm.ListAddress(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(resNUll))
}

func testUpdateNonce(ctx context.Context, t *testing.T, api, apiNoPerm messager.IMessager, allAddrs []address.Address) {
	addrInfos, err := api.ListAddress(ctx)
	assert.NoError(t, err)
	addrNonce := make(map[address.Address]uint64, len(addrInfos))
	for _, addrInfo := range addrInfos {
		addrNonce[addrInfo.Addr] = addrInfo.Nonce
	}
	nonce := uint64(1)
	for _, addr := range allAddrs {
		_, ok := addrNonce[addr]
		if ok {
			latestNonce := addrNonce[addr] + nonce
			assert.NoError(t, api.UpdateNonce(ctx, addr, latestNonce))
			err := apiNoPerm.UpdateNonce(ctx, addr, latestNonce)
			assert.Error(t, err)
			addrInfo, err := api.GetAddress(ctx, addr)
			assert.NoError(t, err)
			assert.Equal(t, latestNonce, addrInfo.Nonce)
		} else {
			assert.NoError(t, api.UpdateNonce(ctx, addr, nonce))
		}
	}
}

func testForbiddenAndActiveAddress(ctx context.Context, t *testing.T, api, apiNoPerm messager.IMessager, allAddrs []address.Address, usedAddrs map[address.Address]struct{}) {
	for _, addr := range allAddrs {
		_, ok := usedAddrs[addr]
		if ok {
			assert.NoError(t, api.ForbiddenAddress(ctx, addr))
			err := apiNoPerm.ForbiddenAddress(ctx, addr)
			assert.Contains(t, err.Error(), "permission deny")
			addrInfo, err := api.GetAddress(ctx, addr)
			assert.NoError(t, err)
			assert.Equal(t, types.AddressStateForbbiden, addrInfo.State)

			// active address
			assert.NoError(t, api.ActiveAddress(ctx, addr))
			err = apiNoPerm.ActiveAddress(ctx, addr)
			assert.Contains(t, err.Error(), "permission deny")
			addrInfo, err = api.GetAddress(ctx, addr)
			assert.NoError(t, err)
			assert.Equal(t, types.AddressStateAlive, addrInfo.State)
		} else {
			assert.NoError(t, api.ForbiddenAddress(ctx, addr))
			assert.NoError(t, api.ActiveAddress(ctx, addr))
		}
	}
}

func testSetSelectMsgNum(ctx context.Context, t *testing.T, api, apiNoPerm messager.IMessager, allAddrs []address.Address, usedAddrs map[address.Address]struct{}) {
	selectNum := uint64(100)
	for _, addr := range allAddrs {
		_, ok := usedAddrs[addr]
		if ok {
			assert.NoError(t, api.SetSelectMsgNum(ctx, addr, selectNum))
			err := apiNoPerm.SetSelectMsgNum(ctx, addr, selectNum)
			assert.Contains(t, err.Error(), "permission deny")
			addrInfo, err := api.GetAddress(ctx, addr)
			assert.NoError(t, err)
			assert.Equal(t, selectNum, addrInfo.SelMsgNum)
		} else {
			assert.NoError(t, api.SetSelectMsgNum(ctx, addr, selectNum))
		}
	}
}

func testSetFeeParams(ctx context.Context, t *testing.T, api, apiNoPerm messager.IMessager, allAddrs []address.Address, usedAddrs map[address.Address]struct{}) {
	gasOverEstimation := 11.25
	gasOverPremium := 44.0
	maxFee := big.NewInt(10001110)
	gasFeeCap := big.NewInt(10001101)
	baseFee := big.NewInt(1010110)
	params := types.AddressSpec{
		GasOverEstimation: gasOverEstimation,
		GasOverPremium:    gasOverPremium,
		MaxFeeStr:         maxFee.String(),
		GasFeeCapStr:      gasFeeCap.String(),
		BaseFeeStr:        baseFee.String(),
	}
	emptyAddrSpec := &types.AddressSpec{}

	checkParams := func(addrInfo *types.Address) {
		assert.Equal(t, gasOverEstimation, addrInfo.GasOverEstimation)
		assert.Equal(t, gasOverPremium, addrInfo.GasOverPremium)
		assert.Equal(t, maxFee, addrInfo.MaxFee)
		assert.Equal(t, gasFeeCap, addrInfo.GasFeeCap)
		assert.Equal(t, baseFee, addrInfo.BaseFee)
	}
	var usedAddr address.Address
	for _, addr := range allAddrs {
		_, ok := usedAddrs[addr]
		params.Address = addr
		if ok {
			usedAddr = addr
			assert.NoError(t, api.SetFeeParams(ctx, &params))
			assert.Contains(t, apiNoPerm.SetFeeParams(ctx, &params).Error(), "permission deny")

			addrInfo, err := api.GetAddress(ctx, addr)
			assert.NoError(t, err)
			checkParams(addrInfo)

			// use empty value
			emptyAddrSpec.Address = addr
			assert.NoError(t, api.SetFeeParams(ctx, emptyAddrSpec))
			addrInfo, err = api.GetAddress(ctx, addr)
			assert.NoError(t, err)
			checkParams(addrInfo)
		} else {
			assert.Error(t, api.SetFeeParams(ctx, &params))
		}
	}

	// set zero value
	params2 := &types.AddressSpec{
		Address:      usedAddr,
		MaxFeeStr:    big.Zero().String(),
		GasFeeCapStr: big.Zero().String(),
		BaseFeeStr:   big.Zero().String(),
	}
	assert.NoError(t, api.SetFeeParams(ctx, params2))
	res, err := api.GetAddress(ctx, params2.Address)
	assert.NoError(t, err)
	assert.Equal(t, gasOverEstimation, gasOverEstimation)
	assert.Equal(t, gasOverPremium, gasOverPremium)
	assert.Equal(t, big.Zero(), res.MaxFee)
	assert.Equal(t, big.Zero(), res.GasFeeCap)
	assert.Equal(t, big.Zero(), res.BaseFee)
}

func testClearUnFillMessage(ctx context.Context, t *testing.T, api, apiNoPerm messager.IMessager, allAddrs []address.Address, addrMsgs map[address.Address][]*types.Message) {
	for _, addr := range allAddrs {
		clearNum, err := api.ClearUnFillMessage(ctx, addr)
		assert.NoError(t, err)
		_, err = apiNoPerm.ClearUnFillMessage(ctx, addr)
		assert.Contains(t, err.Error(), "permission deny")

		msgs := addrMsgs[addr]
		assert.Equal(t, len(msgs), clearNum)

		for _, msg := range msgs {
			res, err := api.GetMessageByUid(ctx, msg.ID)
			assert.NoError(t, err)
			assert.Equal(t, types.FailedMsg, res.State)
		}
	}
}

func testDeleteAddress(ctx context.Context, t *testing.T, api, apiNoPerm messager.IMessager, allAddrs []address.Address, usedAddrs map[address.Address]struct{}) {
	for _, addr := range allAddrs {
		_, ok := usedAddrs[addr]
		if !ok {
			assert.NoError(t, api.DeleteAddress(ctx, addr))
		}
		assert.NoError(t, api.DeleteAddress(ctx, addr))
		err := apiNoPerm.DeleteAddress(ctx, addr)
		assert.Contains(t, err.Error(), "permission deny")

		_, err = api.GetAddress(ctx, addr)
		assert.Contains(t, err.Error(), gorm.ErrRecordNotFound.Error())
	}

	list, err := api.ListAddress(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 0)
}
