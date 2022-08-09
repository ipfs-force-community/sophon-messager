package integration

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/testhelper"
)

const defaultLocalToken = "defaultLocalToken"

func TestAddressAPI(t *testing.T) {
	ctx := context.Background()
	cfg := config.DefaultConfig()
	cfg.API.Address = "/ip4/0.0.0.0/tcp/0"
	cfg.MessageService.SkipPushMessage = true
	cfg.MessageService.WaitingChainHeadStableDuration = 2 * time.Second
	ms, err := mockMessagerServer(ctx, t.TempDir(), cfg)
	assert.NoError(t, err)

	go ms.start(ctx)
	assert.NoError(t, <-ms.appStartErr)

	account := defaultLocalToken
	addrCount := 10
	addrs := testhelper.RandAddresses(t, addrCount)
	assert.NoError(t, ms.walletCli.AddAddress(account, addrs))

	cli, closer, err := newMessagerClient(ctx, ms.port, ms.token)
	assert.NoError(t, err)
	defer closer()

	allAddrs := make([]address.Address, 0, len(addrs))
	for _, addr := range addrs {
		allAddrs = append(allAddrs, testhelper.ResolveAddr(t, addr))
	}

	usedAddrs := make(map[address.Address]struct{})
	msgs := testhelper.NewMessages(addrCount * 2)
	addrMsgNum := make(map[address.Address]int, len(addrs))
	for _, msg := range msgs {
		msg.From = addrs[rand.Intn(addrCount)]
		msg.FromUser = account
		id, err := cli.PushMessageWithId(ctx, msg.ID, &msg.Message, msg.Meta)
		assert.NoError(t, err)
		assert.Equal(t, msg.ID, id)

		tmp := testhelper.ResolveAddr(t, msg.From)
		usedAddrs[tmp] = struct{}{}
		addrMsgNum[tmp]++
	}

	t.Run("test get address and has address", func(t *testing.T) {
		testGetAddressAndHasAddress(ctx, t, cli, allAddrs, usedAddrs)
	})
	t.Run("test wallet has", func(t *testing.T) {
		testWalletHas(ctx, t, cli, allAddrs)
	})
	t.Run("test list address", func(t *testing.T) {
		testListAddress(ctx, t, cli, usedAddrs)
	})
	t.Run("test update nonce", func(t *testing.T) {
		testUpdateNonce(ctx, t, cli, allAddrs)
	})
	t.Run("test forbidden and active address", func(t *testing.T) {
		testForbiddenAndActiveAddress(ctx, t, cli, allAddrs, usedAddrs)
	})
	t.Run("test set select message num", func(t *testing.T) {
		testSetSelectMsgNum(ctx, t, cli, allAddrs, usedAddrs)
	})
	t.Run("test set fee params", func(t *testing.T) {
		testSetFeeParams(ctx, t, cli, allAddrs, usedAddrs)
	})
	t.Run("test clear unfill message", func(t *testing.T) {
		testClearUnFillMessage(ctx, t, cli, allAddrs, addrMsgNum)
	})
	t.Run("test delete address", func(t *testing.T) {
		testDeleteAddress(ctx, t, cli, allAddrs, usedAddrs)
	})

	assert.NoError(t, ms.stop(ctx))
}

func testGetAddressAndHasAddress(ctx context.Context,
	t *testing.T,
	cli messager.IMessager,
	allAddrs []address.Address,
	usedAddrs map[address.Address]struct{}) {
	var err error
	for _, addr := range allAddrs {
		_, ok := usedAddrs[addr]
		addrInfo, getAddrErr := cli.GetAddress(ctx, addr)
		assert.NoError(t, err)

		// test has address
		has, err := cli.HasAddress(ctx, addr)
		assert.NoError(t, err)

		if ok {
			assert.NoError(t, getAddrErr)
			assert.Equal(t, addr, addrInfo.Addr)
			assert.True(t, has)
		} else {
			assert.Error(t, getAddrErr)
			assert.False(t, has)
		}
	}
}

func testWalletHas(ctx context.Context, t *testing.T, cli messager.IMessager, allAddrs []address.Address) {
	for _, addr := range allAddrs {
		has, err := cli.WalletHas(ctx, addr)
		assert.NoError(t, err)
		assert.True(t, has)
	}
}

func testListAddress(ctx context.Context, t *testing.T, cli messager.IMessager, usedAddrs map[address.Address]struct{}) {
	addrInfos, err := cli.ListAddress(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(usedAddrs), len(addrInfos))
	for _, addrInfo := range addrInfos {
		_, ok := usedAddrs[addrInfo.Addr]
		assert.True(t, ok)
		assert.Equal(t, types.AddressStateAlive, addrInfo.State)
	}
}

func testUpdateNonce(ctx context.Context, t *testing.T, cli messager.IMessager, allAddrs []address.Address) {
	addrInfos, err := cli.ListAddress(ctx)
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
			assert.NoError(t, cli.UpdateNonce(ctx, addr, latestNonce))
			addrInfo, err := cli.GetAddress(ctx, addr)
			assert.NoError(t, err)
			assert.Equal(t, latestNonce, addrInfo.Nonce)
		} else {
			assert.NoError(t, cli.UpdateNonce(ctx, addr, nonce))
		}
	}
}

func testForbiddenAndActiveAddress(ctx context.Context, t *testing.T, cli messager.IMessager, allAddrs []address.Address, usedAddrs map[address.Address]struct{}) {
	for _, addr := range allAddrs {
		_, ok := usedAddrs[addr]
		if ok {
			assert.NoError(t, cli.ForbiddenAddress(ctx, addr))
			addrInfo, err := cli.GetAddress(ctx, addr)
			assert.NoError(t, err)
			assert.Equal(t, types.AddressStateForbbiden, addrInfo.State)

			// active address
			assert.NoError(t, cli.ActiveAddress(ctx, addr))
			addrInfo, err = cli.GetAddress(ctx, addr)
			assert.NoError(t, err)
			assert.Equal(t, types.AddressStateAlive, addrInfo.State)
		} else {
			assert.NoError(t, cli.ForbiddenAddress(ctx, addr))
			assert.NoError(t, cli.ActiveAddress(ctx, addr))
		}
	}
}

func testSetSelectMsgNum(ctx context.Context, t *testing.T, cli messager.IMessager, allAddrs []address.Address, usedAddrs map[address.Address]struct{}) {
	selectNum := uint64(100)
	for _, addr := range allAddrs {
		_, ok := usedAddrs[addr]
		if ok {
			assert.NoError(t, cli.SetSelectMsgNum(ctx, addr, selectNum))
			addrInfo, err := cli.GetAddress(ctx, addr)
			assert.NoError(t, err)
			assert.Equal(t, selectNum, addrInfo.SelMsgNum)
		} else {
			assert.NoError(t, cli.SetSelectMsgNum(ctx, addr, selectNum))
		}
	}
}

func testSetFeeParams(ctx context.Context, t *testing.T, cli messager.IMessager, allAddrs []address.Address, usedAddrs map[address.Address]struct{}) {
	gasOverEstimation := 11.25
	gasOverPremium := 44.0
	maxFee := big.NewInt(10001110)
	gasFeeCap := big.NewInt(10001101)

	checkParams := func(addrInfo *types.Address) {
		assert.Equal(t, gasOverEstimation, addrInfo.GasOverEstimation)
		assert.Equal(t, gasOverPremium, addrInfo.GasOverPremium)
		assert.Equal(t, maxFee, addrInfo.MaxFee)
		assert.Equal(t, gasFeeCap, addrInfo.GasFeeCap)
	}
	for _, addr := range allAddrs {
		_, ok := usedAddrs[addr]
		if ok {
			assert.NoError(t, cli.SetFeeParams(ctx, addr, gasOverEstimation, gasOverPremium, maxFee.String(), gasFeeCap.String()))
			addrInfo, err := cli.GetAddress(ctx, addr)
			assert.NoError(t, err)
			checkParams(addrInfo)

			// use empty value
			assert.NoError(t, cli.SetFeeParams(ctx, addr, 0, 0, "", ""))
			addrInfo, err = cli.GetAddress(ctx, addr)
			assert.NoError(t, err)
			checkParams(addrInfo)
		} else {
			assert.Error(t, cli.SetFeeParams(ctx, addr, gasOverEstimation, gasOverPremium, maxFee.String(), gasFeeCap.String()))
		}
	}
}

func testClearUnFillMessage(ctx context.Context, t *testing.T, cli messager.IMessager, allAddrs []address.Address, addrMsgNum map[address.Address]int) {
	for _, addr := range allAddrs {
		num := addrMsgNum[addr]
		clearNum, err := cli.ClearUnFillMessage(ctx, addr)
		assert.NoError(t, err)
		assert.Equal(t, num, clearNum)
	}
}

func testDeleteAddress(ctx context.Context, t *testing.T, cli messager.IMessager, allAddrs []address.Address, usedAddrs map[address.Address]struct{}) {
	for _, addr := range allAddrs {
		_, ok := usedAddrs[addr]
		if !ok {
			assert.NoError(t, cli.DeleteAddress(ctx, addr))
		}
		assert.NoError(t, cli.DeleteAddress(ctx, addr))
		_, err := cli.GetAddress(ctx, addr)
		assert.Error(t, err)
	}

	list, err := cli.ListAddress(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 0)
}
