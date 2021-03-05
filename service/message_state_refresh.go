package service

import (
	"context"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/pkg/chain"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/ipfs-force-community/venus-messager/utils"
	"golang.org/x/xerrors"
)

func (ms *MessageService) GoRefreshMessageState() {
	ms.mutx.Lock()
	defer ms.mutx.Unlock()
	if ms.isStateRefreshTaskRunning {
		ms.log.Infof("task refreshMessageState is running, ignore current invoke!\n")
		return
	}
	ms.isStateRefreshTaskRunning = true

	go func() {
		defer func() {
			ms.mutx.Lock()
			ms.isStateRefreshTaskRunning = false
			ms.mutx.Unlock()
		}()
		if err := ms.DoRefreshMsgsState(); err != nil {
			ms.log.Errorf("DoRefreshMsgsState occurs unexpected err:\n%s\n", err.Error())
		}
	}()
}

func (ms *MessageService) DoRefreshMsgsState() error {
	msgs, err := ms.repo.MessageRepo().ListUnchainedMsgs()
	if err != nil {
		return xerrors.Errorf("listUnchainedMsgs failed:%w", err)
	}

	var total = len(msgs)
	if total == 0 {
		return nil
	}

	const maxParallel, windowSize = 5, 5
	var msgCount = len(msgs)
	var par = utils.NewPar(maxParallel)

	for i := 0; i < msgCount; {
		var from, to = i, i + windowSize
		if to > msgCount {
			to = msgCount
		}
		par.Go(func() error {
			return ms.batchRefreshMsgsState(msgs[from:to])
		})
		i = to
	}

	if errs := par.Wait(); errs.Len() != 0 {
		return errs
	}
	return nil
}

func (ms *MessageService) batchRefreshMsgsState(msgs []*types.Message) error {
	var multiErr, err = &utils.MultiError{}, error(nil)

	for _, msg := range msgs {
		if err = ms.refreshSingalMsgState(msg); err != nil {
			multiErr.AddError(err)
		}
	}

	if multiErr.Len() != 0 {
		return multiErr.ERR()
	}
	return nil
}

func (ms *MessageService) refreshSingalMsgState(msg *types.Message) error {
	var fromEpoch = abi.ChainEpoch(0)
	var msgLokup *chain.MsgLookup
	var err error

	if msgLokup, err = ms.nodeClient.StateSearchMsg(
		context.TODO(), msg.UnsingedCid()); err != nil {
		return xerrors.Errorf("SearchMsgLimited(%s, %d) failed:%w",
			msg.SignedCid().String(), fromEpoch, err)
	}

	if msgLokup == nil {
		return nil
	}

	msg.Epoch = uint64(msgLokup.Height)
	msg.Receipt = &msgLokup.Receipt

	if _, err = ms.repo.MessageRepo().UpdateMessageReceipt(msg); err != nil {
		return xerrors.Errorf("UpdateMessageReceipt(uid:%s, cid:%s) failed:%w",
			msg.Uid, msg.UnsingedCid().String(), err)
	}
	return nil
}
