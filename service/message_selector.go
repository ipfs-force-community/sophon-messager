package service

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/filecoin-project/go-address"
	venusTypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/ipfs-force-community/venus-wallet/core"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

type MessageSelector struct {
	repo           repo.Repo
	log            *logrus.Logger
	cfg            *config.MessageServiceConfig
	nodeClient     *NodeClient
	addressService *AddressService
}

func NewMessageSelector(repo repo.Repo, log *logrus.Logger, cfg *config.MessageServiceConfig, nodeClient *NodeClient, addressService *AddressService) *MessageSelector {
	return &MessageSelector{repo: repo, log: log, cfg: cfg, nodeClient: nodeClient, addressService: addressService}
}

func (messageSelector *MessageSelector) SelectMessage(ctx context.Context, ts *venusTypes.TipSet) ([]*types.Message, []*types.Message, []*venusTypes.SignedMessage, []*types.Address, error) {
	addrList, err := messageSelector.addressService.ListAddress(ctx)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	//sort by addr weight
	sort.Slice(addrList, func(i, j int) bool {
		return addrList[i].Weight < addrList[j].Weight
	})

	var selectMsg []*types.Message
	var expireMsgs []*types.Message
	var toPushMessage []*venusTypes.SignedMessage
	var modifyAddrs []*types.Address

	var lk sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)
	wg.Add(len(addrList))
	for _, addr := range addrList {
		go func(addr *types.Address) {
			sem <- struct{}{}
			defer func() {
				wg.Done()
				<-sem
			}()

			addrSelectMsg, addrExpireMsgs, addrToPushMessage, err := messageSelector.selectAddrMessage(ctx, addr, ts)
			if err != nil {
				messageSelector.log.Errorf("select message of %s fail %v", addr.Addr, err)
				return
			}
			lk.Lock()
			defer lk.Unlock()

			expireMsgs = append(expireMsgs, addrExpireMsgs...)
			toPushMessage = append(toPushMessage, addrToPushMessage...)
			if len(addrSelectMsg) > 0 {
				selectMsg = append(selectMsg, addrSelectMsg...)
				modifyAddrs = append(modifyAddrs, addr)
			}
		}(addr)
	}

	wg.Wait()

	return selectMsg, expireMsgs, toPushMessage, modifyAddrs, nil
}

func (messageSelector *MessageSelector) selectAddrMessage(ctx context.Context, addr *types.Address, ts *venusTypes.TipSet) ([]*types.Message, []*types.Message, []*venusTypes.SignedMessage, error) {
	maxAllowPendingMessage := uint64(50)
	var toPushMessage []*venusTypes.SignedMessage

	addrInfo, exit := messageSelector.addressService.GetAddressInfo(addr.Addr)
	if !exit {
		return nil, nil, nil, xerrors.Errorf("no wallet cliet of address %s", addr.Addr)
	}

	mAddr, err := address.NewFromString(addr.Addr)
	if err != nil {
		return nil, nil, nil, xerrors.Errorf("addr format error %s", addr.Addr)
	}

	//判断是否需要推送消息
	actor, err := messageSelector.nodeClient.StateGetActor(ctx, mAddr, ts.Key())
	if err != nil {
		return nil, nil, nil, xerrors.Errorf("actor of address %s not found", mAddr)
	}

	if actor.Nonce > addr.Nonce {
		messageSelector.log.Warnf("%s nonce in db %d is smaller than nonce on chain %d, update to latest", addr.Addr, addr.Nonce, actor.Nonce)
		addr.Nonce = actor.Nonce
		addr.UpdatedAt = time.Now()
		_, err := messageSelector.repo.AddressRepo().SaveAddress(ctx, addr)
		if err != nil {
			return nil, nil, nil, xerrors.Errorf("update address %s nonce fail", addr.Addr)
		}
	}
	//todo push sigined but not onchain message, when to resend message
	filledMessage, err := messageSelector.repo.MessageRepo().ListFilledMessageByAddress(mAddr)
	if err != nil {
		messageSelector.log.Warnf("list filled message %v", err)
	}
	for _, msg := range filledMessage {
		toPushMessage = append(toPushMessage, &venusTypes.SignedMessage{
			Message:   msg.UnsignedMessage,
			Signature: *msg.Signature,
		})
	}

	//消息排序
	messages, err := messageSelector.repo.MessageRepo().ListUnChainMessageByAddress(mAddr)
	if err != nil {
		return nil, nil, nil, xerrors.Errorf("list %s unpackage message error %v", mAddr, err)
	}
	messages, expireMsgs := messageSelector.excludeExpire(ts, messages)

	//sign new message
	nonceGap := addr.Nonce - actor.Nonce
	if nonceGap > maxAllowPendingMessage {
		messageSelector.log.Infof("%s there are %d message not to be package ", addr.Addr, nonceGap)
		return nil, expireMsgs, toPushMessage, nil
	}
	selectCount := maxAllowPendingMessage - nonceGap

	//todo 如何筛选
	if len(messages) == 0 {
		messageSelector.log.Infof("%s have no message", addr.Addr)
		return nil, expireMsgs, toPushMessage, nil
	}

	var count = uint64(0)
	var selectMsg []*types.Message
	for _, msg := range messages {
		//分配nonce
		msg.Nonce = addr.Nonce

		//todo 估算gas, spec怎么做？
		//通过配置影响 maxfee
		newMsg, err := messageSelector.nodeClient.GasEstimateMessageGas(ctx, msg.VMMessage(), &venusTypes.MessageSendSpec{MaxFee: msg.Meta.MaxFee}, ts.Key())
		if err != nil {
			if strings.Contains(err.Error(), "exit SysErrSenderStateInvalid(2)") {
				// SysErrSenderStateInvalid(2))
				messageSelector.log.Errorf("message %s estimate message fail %v break address %s", msg.ID, err, addr.Addr)
				break
			}
			messageSelector.log.Errorf("message %s estimate message fail %v, try to next message", msg.ID, err)

			continue
		}
		msg.GasFeeCap = newMsg.GasFeeCap
		msg.GasPremium = newMsg.GasPremium
		msg.GasLimit = newMsg.GasLimit

		unsignedCid := msg.UnsignedMessage.Cid()
		msg.UnsignedCid = &unsignedCid
		//签名
		data, err := msg.UnsignedMessage.ToStorageBlock()
		if err != nil {
			messageSelector.log.Errorf("calc message unsigned message id %s fail %v", msg.ID.String(), err)
			continue
		}
		sig, err := addrInfo.WalletClient.WalletSign(ctx, mAddr, unsignedCid.Bytes(), core.MsgMeta{
			Type:  core.MTChainMsg,
			Extra: data.RawData(),
		})
		if err != nil {
			//todo client net crash?
			messageSelector.log.Errorf("wallet sign failed %s fail %v", msg.ID.String(), err)
			continue
		}

		msg.Signature = sig
		//state
		msg.State = types.FillMsg

		signedMsg := venusTypes.SignedMessage{
			Message:   msg.UnsignedMessage,
			Signature: *msg.Signature,
		}

		signedCid := signedMsg.Cid()
		msg.SignedCid = &signedCid

		selectMsg = append(selectMsg, msg)
		addr.Nonce++
		count++
		if count >= selectCount {
			break
		}
	}

	messageSelector.log.Infof("address %s select message %d max nonce %d", addr.Addr, len(selectMsg), addr.Nonce)
	return selectMsg, expireMsgs, toPushMessage, nil
}

func (messageSelector *MessageSelector) excludeExpire(ts *venusTypes.TipSet, msgs []*types.Message) ([]*types.Message, []*types.Message) {
	//todo check whether message is expired
	var result []*types.Message
	var expireMsg []*types.Message
	for _, msg := range msgs {
		if msg.Meta.ExpireEpoch != 0 && msg.Meta.ExpireEpoch <= ts.Height() {
			//expire
			msg.State = types.ExpireMsg
			expireMsg = append(expireMsg, msg)
			continue
		}
		result = append(result, msg)
	}
	return result, expireMsg
}
