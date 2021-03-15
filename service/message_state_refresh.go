package service

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/filecoin-project/go-address"

	"github.com/filecoin-project/go-state-types/abi"
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

	return ms.repo.Transaction(func(txRepo repo.TxRepo) error {
		var revertMsgs map[cid.Cid]struct{}
		if len(h.revert) != 0 {
			var err error
			revertMsgs, err = ms.processRevertHead(ctx, txRepo, h)
			if err != nil {
				return err
			}
		}

		for _, ts := range h.apply {
			height := ts.Height()
			if !ts.Defined() {
				continue
			}
			if err := ms.processBlockParentMessages(ctx, txRepo, ts.At(0).Cid(), height, revertMsgs); err != nil {
				return xerrors.Errorf("process block failed, block id: %s %v", ts.At(0).Cid().String(), err)
			}
			ms.tsCache.AddTs(&tipsetFormat{Key: ts.String(), Height: uint64(height)})
		}
		if len(h.apply) > 0 {
			ms.tsCache.CurrHeight = uint64(h.apply[0].Height())
		}

		if err := ms.storeTipset(); err != nil {
			ms.log.Errorf("store tipset info failed: %v", err)
		}
		ms.triggerPush <- h.apply[0]
		return nil
	})
}

func (ms *MessageService) processRevertHead(ctx context.Context, txRepo repo.TxRepo, h *headChan) (map[cid.Cid]struct{}, error) {
	var c cid.Cid
	revertMsgs := make(map[cid.Cid]struct{})
	for _, tipset := range h.revert {
		for _, block := range tipset.Cids() {
			msgs, err := ms.nodeClient.ChainGetBlockMessages(ctx, block)
			if err != nil {
				return nil, xerrors.Errorf("get block message failed %v", err)
			}

			for _, msg := range msgs.BlsMessages {
				if _, ok := ms.addressService.addrInfo[msg.From.String()]; ok {
					c = msg.Cid()
					revertMsgs[c] = struct{}{}
					if err := txRepo.MessageRepo().UpdateMessageStateByCid(c.String(), types.UnFillMsg); err != nil {
						return nil, err
					}
					if err := ms.messageState.UpdateMessageStateByCid(c.String(), types.UnFillMsg); err != nil {
						ms.log.Errorf("update message state failed, cid: %s error: %v", c.String(), err)
					}
				}

			}

			for _, msg := range msgs.SecpkMessages {
				if _, ok := ms.addressService.addrInfo[msg.Message.From.String()]; ok {
					c = msg.Message.Cid()
					revertMsgs[c] = struct{}{}
					if err := txRepo.MessageRepo().UpdateMessageStateByCid(c.String(), types.UnFillMsg); err != nil {
						return nil, err
					}
					if err := ms.messageState.UpdateMessageStateByCid(c.String(), types.UnFillMsg); err != nil {
						ms.log.Errorf("update message state failed, cid: %s error: %v", c.String(), err)
					}
				}
			}
		}
	}

	return revertMsgs, nil
}

func (ms *MessageService) processBlockParentMessages(ctx context.Context, txRepo repo.TxRepo, bcid cid.Cid, height abi.ChainEpoch, revertMsgs map[cid.Cid]struct{}) error {
	msgs, err := ms.nodeClient.ChainGetParentMessages(ctx, bcid)
	if err != nil {
		return xerrors.Errorf("get parent message failed %w", err)
	}

	receipts, err := ms.nodeClient.ChainGetParentReceipts(ctx, bcid)
	if err != nil {
		return xerrors.Errorf("get parent Receipt failed %w", err)
	}

	if len(msgs) != len(receipts) {
		return xerrors.Errorf("messages not match receipts, %d != %d", len(msgs), len(receipts))
	}

	gapNonce := make(map[address.Address]uint64)
	for i := range receipts {
		msg := msgs[i].Message
		if addrInfo, ok := ms.addressService.addrInfo[msg.From.String()]; ok {
			cidStr := msg.Cid().String()
			if _, err = txRepo.MessageRepo().UpdateMessageReceipt(cidStr, receipts[i], height, types.OnChainMsg); err != nil {
				return xerrors.Errorf("update message receipt failed, cid:%s failed:%v", cidStr, err)
			}
			if _, ok := revertMsgs[msgs[i].Cid]; ok {
				if err := ms.messageState.UpdateMessageStateByCid(msg.Cid().String(), types.OnChainMsg); err != nil {
					ms.log.Errorf("update message state failed, cid: %s error: %v", msg.Cid().String(), err)
				}
			}
			if addrInfo.Nonce < msg.Nonce {
				if nonce, ok := gapNonce[msg.From]; ok && nonce >= msg.Nonce {
					continue
				}
				gapNonce[msg.From] = msg.Nonce
			}
		}
	}

	for addr, nonce := range gapNonce {
		if err := ms.addressService.StoreNonce(addr.String(), nonce); err != nil {
			return err
		}
		ms.addressService.SetNonce(addr.String(), nonce)
	}

	return nil
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
