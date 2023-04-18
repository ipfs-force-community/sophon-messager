package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"go.opencensus.io/stats"

	venustypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"

	"github.com/filecoin-project/venus-messager/metrics"
	"github.com/filecoin-project/venus-messager/models/repo"
)

var msgStateLog = logging.Logger("msg-state")

func (ms *MessageService) refreshMessageState(ctx context.Context) {
	go func() {
		for {
			select {
			case h := <-ms.headChans:
				// 跳过这个检查可以更精准的推送，但是会增加系统负担
				/*	if len(h.apply) == 1 && len(h.revert) == 1 {
					ms.tsCache.AddTs(&tipsetFormat{Key: h.apply[0].Key().String(), Height: int64(h.apply[0].Height())})
					stateRefreshlog.Warnf("revert at same height %d just update cache and skip process %s", h.apply[0].Height(), h.apply[0].String())
					ms.triggerPush <- h.apply[0]
					continue
				}*/
				msgStateLog.Infof("start refresh message state, apply %d, revert %d", len(h.apply), len(h.revert))
				now := time.Now()
				if err := ms.doRefreshMessageState(ctx, h); err != nil {
					h.done <- err
					msgStateLog.Errorf("refresh message occurs unexpected error %v", err)
					continue
				}
				h.done <- nil
				msgStateLog.Infof("end refresh message state, spent %d 'ms'", time.Since(now).Milliseconds())
			case <-ctx.Done():
				msgStateLog.Warnf("stop refresh message state: %v", ctx.Err())
				return
			}
		}
	}()
}

func (ms *MessageService) doRefreshMessageState(ctx context.Context, h *headChan) error {
	if len(h.apply) == 0 {
		msgStateLog.Infof("apply is empty")
		return nil
	}

	revertMsgs, err := ms.processRevertHead(ctx, h)
	if err != nil {
		return err
	}

	applyMsgs, err := ms.processBlockParentMessages(ctx, h.apply)
	if err != nil {
		return fmt.Errorf("process apply failed %v", err)
	}

	// update db
	replaceMsg, invalidMsgs, err := ms.updateMessageState(ctx, applyMsgs, revertMsgs)
	if err != nil {
		return err
	}

	ms.tsCache.CurrHeight = int64(h.apply[0].Height())
	ms.tsCache.Add(h.apply...)
	if err := ms.tsCache.Save(ms.fsRepo.TipsetFile()); err != nil {
		msgStateLog.Errorf("store tipsetkey failed %v", err)
	}

	msgStateLog.Infof("process block %d, revert %d message, apply %d message, replaced %d message", ms.tsCache.CurrHeight, len(revertMsgs), len(applyMsgs)-len(invalidMsgs), len(replaceMsg))

	if ms.preCancel != nil {
		ms.preCancel()
	}
	if !h.isReconnect { // reconnect do not push messager avoid wrong gas estimate
		var triggerCtx context.Context
		triggerCtx, ms.preCancel = context.WithCancel(context.Background())
		go ms.delayTrigger(triggerCtx, h.apply[0])
	}
	return nil
}

func (ms *MessageService) updateMessageState(ctx context.Context, applyMsgs []applyMessage, revertMsgs map[cid.Cid]struct{}) (map[string]*types.Message, map[cid.Cid]struct{}, error) {
	replaceMsg := make(map[string]*types.Message)
	invalidMsgs := make(map[cid.Cid]struct{})
	return replaceMsg, invalidMsgs, ms.repo.Transaction(func(txRepo repo.TxRepo) error {
		for cid := range revertMsgs {
			if err := txRepo.MessageRepo().UpdateMessageInfoByCid(cid.String(), &venustypes.MessageReceipt{ExitCode: -1},
				abi.ChainEpoch(0), types.FillMsg, venustypes.EmptyTSK); err != nil {
				return err
			}
		}

		for _, msg := range applyMsgs {
			// 两个 `nonce` 都为 `0` 的消息，第一条消息预估gas失败了，第二条消息成功上链，
			// 若只按 `from` 和 `nonce` 查询，查到的是第一条消息，这样第二条消息一直是 `FillMsg`
			localMsg, err := txRepo.MessageRepo().GetMessageByFromNonceAndState(msg.msg.From, msg.msg.Nonce, types.FillMsg)
			if err != nil {
				msgStateLog.Warnf("msg %s not exist in local db maybe address %s send out of messager", msg.signedCID, msg.msg.From)
				invalidMsgs[msg.signedCID] = struct{}{}
				continue
			}
			if localMsg.SignedCid != nil && !(*localMsg.SignedCid).Equals(msg.signedCID) {
				msgStateLog.Warnf("replace message old msg cid %s, new msg cid %s, id %s", localMsg.SignedCid, msg.signedCID, localMsg.ID)
				// replace msg
				localMsg.State = types.NonceConflictMsg
				localMsg.Receipt = msg.receipt
				localMsg.Height = int64(msg.height)
				localMsg.TipSetKey = msg.tsk
				if err = txRepo.MessageRepo().UpdateMessage(localMsg); err != nil {
					return fmt.Errorf("update message receipt failed, cid:%s failed:%v", msg.signedCID, err)
				}
				replaceMsg[localMsg.ID] = localMsg
			} else {
				if err = txRepo.MessageRepo().UpdateMessageInfoByCid(msg.msg.Cid().String(), msg.receipt, msg.height, types.OnChainMsg, msg.tsk); err != nil {
					return fmt.Errorf("update message receipt failed, cid:%s failed:%v", msg.msg.Cid(), err)
				}
			}
		}
		return nil
	})
}

// delayTrigger wait for stable ts
func (ms *MessageService) delayTrigger(ctx context.Context, ts *venustypes.TipSet) {
	select {
	case <-time.After(ms.fsRepo.Config().MessageService.WaitingChainHeadStableDuration):
		ds := time.Now().Unix() - int64(ts.MinTimestamp())
		stats.Record(ctx, metrics.ChainHeadStableDelay.M(ds))
		stats.Record(ctx, metrics.ChainHeadStableDuration.M(ds))
		ms.triggerPush <- ts
		return
	case <-ctx.Done():
		return
	}
}

func (ms *MessageService) processRevertHead(ctx context.Context, h *headChan) (map[cid.Cid]struct{}, error) {
	revertMsgs := make(map[cid.Cid]struct{})

	var msgCIDs []string
	for _, ts := range h.revert {
		msgs, err := ms.repo.MessageRepo().ListChainMessageByHeight(ts.Height())
		if err != nil {
			return nil, fmt.Errorf("found filled message at height %d error %v", ts.Height(), err)
		}

		addrs := ms.addressService.ActiveAddresses(ctx)
		for _, msg := range msgs {
			if _, ok := addrs[msg.From]; ok && msg.UnsignedCid != nil {
				revertMsgs[*msg.UnsignedCid] = struct{}{}
				msgCIDs = append(msgCIDs, (*msg.UnsignedCid).String())
			}
		}

		if len(msgCIDs) > 0 {
			log.Infof("revert %d messages %v at height %d", len(msgCIDs), strings.Join(msgCIDs, ","), ts.Height())
			msgCIDs = msgCIDs[:0]
		}
	}

	return revertMsgs, nil
}

type applyMessage struct {
	signedCID cid.Cid
	msg       *venustypes.Message
	height    abi.ChainEpoch
	tsk       venustypes.TipSetKey
	receipt   *venustypes.MessageReceipt
}

func (ms *MessageService) processBlockParentMessages(ctx context.Context, apply []*venustypes.TipSet) ([]applyMessage, error) {
	var applyMsgs []applyMessage
	var msgCIDs []string
	addrs := ms.addressService.ActiveAddresses(ctx)
	for _, ts := range apply {
		bcid := ts.At(0).Cid()
		msgs, err := ms.nodeClient.ChainGetParentMessages(ctx, bcid)
		if err != nil {
			return nil, fmt.Errorf("got parent message failed %w", err)
		}

		receipts, err := ms.nodeClient.ChainGetParentReceipts(ctx, bcid)
		if err != nil {
			return nil, fmt.Errorf("got parent receipt failed %w", err)
		}

		if len(msgs) != len(receipts) {
			return nil, fmt.Errorf("messages not match receipts, %d != %d", len(msgs), len(receipts))
		}

		pts, err := ms.nodeClient.ChainGetTipSet(ctx, ts.Parents())
		if err != nil {
			return nil, fmt.Errorf("got parent ts failed: %v", err)
		}

		for i := range receipts {
			msg := msgs[i].Message
			if _, ok := addrs[msg.From]; ok {
				applyMsgs = append(applyMsgs, applyMessage{
					height:    pts.Height(),
					tsk:       pts.Key(),
					receipt:   receipts[i],
					msg:       msg,
					signedCID: msgs[i].Cid,
				})
				msgCIDs = append(msgCIDs, msg.Cid().String())
			}

		}
		if len(msgCIDs) > 0 {
			log.Debugf("apply %d messages %v at height %d", len(msgCIDs), strings.Join(msgCIDs, ","), pts.Height())
			msgCIDs = msgCIDs[:0]
		}
	}
	return applyMsgs, nil
}
