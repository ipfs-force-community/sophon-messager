package service

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs/go-cid"
	"golang.org/x/xerrors"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

func (ms *MessageService) refreshMessageState(ctx context.Context) {
	go func() {
		for {
			select {
			case h := <-ms.headChans:
				ms.log.Info("start refresh message state")
				now := time.Now()
				if err := ms.doRefreshMessageState(ctx, h); err != nil {
					ms.log.Errorf("doRefreshMessageState occurs unexpected err:\n%v\n", err)
				}
				ms.log.Infof("end refresh message state, cost %d 'ms' ", time.Since(now).Milliseconds())
			case <-ctx.Done():
				ms.log.Warnf("context error: %v", ctx.Err())
				return
			}
		}
	}()
}

func (ms *MessageService) doRefreshMessageState(ctx context.Context, h *headChan) error {
	if len(h.apply) == 0 && len(h.revert) == 0 {
		return nil
	}

	var revertMsgs map[cid.Cid]struct{}
	var err error
	var tsList tipsetList

	if len(h.revert) != 0 {
		revertMsgs, err = ms.processRevertHead(ctx, h)
		if err != nil {
			ms.failedHeads = append(ms.failedHeads, failedHead{headChan: headChan{h.apply, h.revert}, Time: time.Now()})
			return err
		}
	}

	pendingMsgs := make([]pendingMessage, 0)
	nonceGap := make(map[address.Address]uint64, len(ms.addressService.addrInfo))
	for _, ts := range h.apply {
		height := ts.Height()
		if !ts.Defined() {
			continue
		}
		pendingMsgs, nonceGap, err = ms.processBlockParentMessages(ctx, ts.At(0).Cid(), height, pendingMsgs, nonceGap)
		if err != nil {
			return xerrors.Errorf("process block failed, block id: %s %v", ts.At(0).Cid().String(), err)
		}
		tsList = append(tsList, &tipsetFormat{Key: ts.String(), Height: uint64(height)})
	}

	// update db
	err = ms.repo.Transaction(func(txRepo repo.TxRepo) error {
		for _, msg := range pendingMsgs {
			if _, err = txRepo.MessageRepo().UpdateMessageReceipt(msg.cid.String(), msg.receipt, msg.height, types.OnChainMsg); err != nil {
				return xerrors.Errorf("update message receipt failed, cid:%s failed:%v", msg.cid.String(), err)
			}
			if _, ok := revertMsgs[msg.cid]; ok {
				delete(revertMsgs, msg.cid)
			}
		}
		for cid := range revertMsgs {
			if err := txRepo.MessageRepo().UpdateMessageStateByCid(cid.String(), types.UnFillMsg); err != nil {
				return err
			}
		}
		for addr, nonce := range nonceGap {
			if err := ms.addressService.StoreNonce(addr.String(), nonce); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		ms.failedHeads = append(ms.failedHeads, failedHead{headChan: headChan{h.apply, h.revert}, Time: time.Now()})
		return err
	}

	// update cache
	for _, msg := range pendingMsgs {
		if err := ms.messageState.UpdateMessageStateByCid(msg.cid.String(), types.OnChainMsg); err != nil {
			ms.log.Errorf("update message state failed, cid: %s error: %v", msg.cid.String(), err)
		}
	}
	for cid := range revertMsgs {
		if err := ms.messageState.UpdateMessageStateByCid(cid.String(), types.UnFillMsg); err != nil {
			ms.log.Errorf("update message state failed, cid: %s error: %v", cid.String(), err)
		}
	}
	for addr, nonce := range nonceGap {
		ms.addressService.SetNonce(addr.String(), nonce)
	}

	if len(h.apply) > 0 {
		ms.tsCache.CurrHeight = uint64(h.apply[0].Height())
		ms.tsCache.AddTs(tsList...)
		if err := ms.storeTipset(); err != nil {
			ms.log.Errorf("store tipset info failed: %v", err)
		}
	}

	ms.triggerPush <- h.apply[0]

	return nil
}

func (ms *MessageService) processRevertHead(ctx context.Context, h *headChan) (map[cid.Cid]struct{}, error) {
	revertMsgs := make(map[cid.Cid]struct{})
	for _, tipset := range h.revert {
		if tipset.Defined() {
			msgs, err := ms.nodeClient.ChainGetParentMessages(ctx, tipset.At(0).Cid())
			if err != nil {
				return nil, xerrors.Errorf("get block message failed %v", err)
			}

			for _, msg := range msgs {
				if _, ok := ms.addressService.addrInfo[msg.Message.From.String()]; ok {
					revertMsgs[msg.Message.Cid()] = struct{}{}
				}

			}
		}
	}

	return revertMsgs, nil
}

type pendingMessage struct {
	cid     cid.Cid
	height  abi.ChainEpoch
	receipt *venustypes.MessageReceipt
}

func (ms *MessageService) processBlockParentMessages(ctx context.Context,
	bcid cid.Cid,
	height abi.ChainEpoch,
	pendingMsgs []pendingMessage,
	nonceGap map[address.Address]uint64) ([]pendingMessage, map[address.Address]uint64, error) {
	msgs, err := ms.nodeClient.ChainGetParentMessages(ctx, bcid)
	if err != nil {
		return nil, nil, xerrors.Errorf("get parent message failed %w", err)
	}

	receipts, err := ms.nodeClient.ChainGetParentReceipts(ctx, bcid)
	if err != nil {
		return nil, nil, xerrors.Errorf("get parent Receipt failed %w", err)
	}

	if len(msgs) != len(receipts) {
		return nil, nil, xerrors.Errorf("messages not match receipts, %d != %d", len(msgs), len(receipts))
	}

	for i := range receipts {
		msg := msgs[i].Message
		if addrInfo, ok := ms.addressService.addrInfo[msg.From.String()]; ok {
			pendingMsgs = append(pendingMsgs, pendingMessage{
				cid:     msg.Cid(),
				height:  height,
				receipt: receipts[i],
			})
			if addrInfo.Nonce < msg.Nonce {
				if nonce, ok := nonceGap[msg.From]; ok && nonce >= msg.Nonce {
					continue
				}
				nonceGap[msg.From] = msg.Nonce
			}
		}
	}

	return pendingMsgs, nonceGap, nil
}

type tipsetFormat struct {
	Key    string
	Height uint64
}

func (ms *MessageService) storeTipset() error {
	ms.tsCache.ReduceTs()

	return updateTipsetFile(ms.cfg.TipsetFilePath, ms.tsCache)
}

type tipsetList []*tipsetFormat

func (tl tipsetList) Len() int {
	return len(tl)
}

func (tl tipsetList) Swap(i, j int) {
	tl[i], tl[j] = tl[j], tl[i]
}

func (tl tipsetList) Less(i, j int) bool {
	return tl[i].Height > tl[j].Height
}

func readTipsetFile(filePath string) (*TipsetCache, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if len(b) < 3 { // skip empty content
		return &TipsetCache{
			Cache:      map[uint64]*tipsetFormat{},
			CurrHeight: 0,
		}, nil
	}
	var tsCache TipsetCache
	if err := json.Unmarshal(b, &tsCache); err != nil {
		return nil, err
	}

	return &tsCache, nil
}

// original data will be cleared
func updateTipsetFile(filePath string, tsCache *TipsetCache) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := json.Marshal(tsCache)
	if err != nil {
		return err
	}
	_, err = file.Write(b)

	return err
}
