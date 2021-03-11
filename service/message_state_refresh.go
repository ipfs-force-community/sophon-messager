package service

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/filecoin-project/go-state-types/abi"
	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/hashicorp/go-multierror"
	"github.com/ipfs/go-cid"
	"golang.org/x/xerrors"

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
			default:
			}
		}
	}()
}

func (ms *MessageService) doRefreshMessageState(ctx context.Context, h *headChan) error {
	if len(h.apply) == 0 && len(h.revert) == 0 {
		return nil
	}

	var errs *multierror.Error
	if len(h.revert) != 0 {
		if err := ms.recordRevertMsgs(ctx, h); err != nil {
			errs = multierror.Append(errs, xerrors.Errorf("process revert failed %v", err))
		}

		tsList := ms.ListTs()
		sort.Sort(tsList)

		minHeight := h.revert[0].Height()
		earliestTs := h.revert[0]
		for i, ts := range h.revert {
			if i == 0 {
				continue
			}
			if ts.Height() < minHeight {
				minHeight = ts.Height()
				earliestTs = h.revert[i]
			}
		}
		for i, ts := range tsList {
			if ts.Height == uint64(minHeight) {
				if isEqual(ts, earliestTs) {
					if len(tsList) <= i+1 {
						ms.ClearTs()
						if err := resetTipsetFile(ms.cfg.TipsetFilePath); err != nil {
							return err
						}
					}
					ms.RemoveTs(tsList[:i+1])
					if err := updateTipsetFile(ms.cfg.TipsetFilePath, tsList[i+1:]); err != nil {
						return err
					}
					break
				}
			}
		}
	}

	for i, ts := range h.apply {
		height := ts.Height()
		if err := ms.processOneBlock(ctx, ts.At(0).Cid(), height); err != nil {
			errs = multierror.Append(errs, xerrors.Errorf("block id: %s %v", ts.At(0).Cid().String(), err))
		}

		if err := ms.storeTipset(h.apply[i]); err != nil {
			ms.log.Errorf("store tipset info failed: %v", err)
		}
	}

	return errs.ErrorOrNil()
}

func (ms *MessageService) recordRevertMsgs(ctx context.Context, h *headChan) error {
	revertMsgs := make([]cid.Cid, 0, 0)
	for _, tipset := range h.revert {
		for _, block := range tipset.Cids() {
			msgs, err := ms.nodeClient.ChainGetBlockMessages(ctx, block)
			if err != nil {
				return xerrors.Errorf("get block message failed %v", err)
			}

			for _, msg := range msgs.BlsMessages {
				if _, ok := ms.addressService.addrInfo[msg.From.String()]; ok {
					revertMsgs = append(revertMsgs, msg.Cid())
				}

			}

			for _, msg := range msgs.SecpkMessages {
				if _, ok := ms.addressService.addrInfo[msg.Message.From.String()]; ok {
					revertMsgs = append(revertMsgs, msg.Cid())
				}
			}
		}
	}

	// update message state
	for _, cid := range revertMsgs {
		ms.messageState.UpdateMessageStateAndReceipt(cid.String(), types.Revert, nil)
		if err := ms.repo.MessageRepo().UpdateMessageStateByCid(cid.String(), types.Revert); err != nil {
			ms.log.Errorf("update message state failed, cid: %s, error: %v", cid.String(), err)
		}
	}

	return nil
}

func (ms *MessageService) processOneBlock(ctx context.Context, bcid cid.Cid, height abi.ChainEpoch) error {
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

	var errs *multierror.Error
	for i := range receipts {
		msg := msgs[i].Message
		if _, ok := ms.addressService.addrInfo[msg.From.String()]; ok {
			cidStr := msgs[i].Cid.String()
			if _, err = ms.repo.MessageRepo().UpdateMessageReceipt(cidStr, receipts[i], height, types.OnChain); err != nil {
				errs = multierror.Append(errs, xerrors.Errorf("cid:%s failed:%v", cidStr, err))
			}

			ms.messageState.UpdateMessageStateAndReceipt(cidStr, types.OnChain, nil)
		}
	}

	return errs.ErrorOrNil()
}

type tipsetFormat struct {
	Cid    []string
	Height uint64
}

func (ms *MessageService) storeTipset(ts *venustypes.TipSet) error {
	if ms.ExistTs(uint64(ts.Height())) {
		ms.log.Warnf("exist same data, height: %d", ts.Height())
		return nil
	}
	cids := ts.Cids()
	format := tipsetFormat{
		Cid:    make([]string, len(cids)),
		Height: uint64(ts.Height()),
	}

	for i := range cids {
		format.Cid[i] = cids[i].String()
	}
	ms.AddTs(&format)

	return appendTipsetFile(ms.cfg.TipsetFilePath, format)
}

func appendTipsetFile(filePath string, ts tipsetFormat) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	b, err := json.Marshal(ts)
	if err != nil {
		return err
	}
	w.WriteString(string(b) + "\n")

	return w.Flush()
}

func resetTipsetFile(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(""); err != nil {
		return err
	}

	return nil
}

// original data will be cleared
func updateTipsetFile(filePath string, lists tipsetList) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	g := multierror.Group{}
	c := make(chan struct{}, 3)
	for _, ts := range lists {
		c <- struct{}{}
		tmp := *ts
		g.Go(func() error {
			b, err := json.Marshal(tmp)
			if err != nil {
				return err
			}
			_, err = writer.WriteString(string(b) + "\n")
			<-c
			return err
		})
	}

	errs := g.Wait()
	close(c)
	errs = multierror.Append(errs, writer.Flush())

	return errs.ErrorOrNil()
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

func (tl tipsetList) Map() map[uint64]*tipsetFormat {
	m := make(map[uint64]*tipsetFormat, len(tl))
	for _, t := range tl {
		m[t.Height] = &tipsetFormat{
			Cid:    t.Cid,
			Height: t.Height,
		}
	}

	return m
}

const bufSize = 1024
const processNum = 3

func readTipsetFromFile(filePath string) (tipsetList, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		return tipsetList{}, err
	}

	c := make(chan struct{}, processNum)
	defer func() {
		close(c)
	}()

	reader := bufio.NewReader(file)
	buf := make([]byte, bufSize)

	var tsList tipsetList

	handleData := func(b []byte) error {
		lines := strings.Split(string(b), "\n")
		for _, l := range lines {
			var ts tipsetFormat
			if len(l) == 0 {
				continue
			}
			if err := json.Unmarshal([]byte(l), &ts); err != nil {
				return err
			}

			tsList = append(tsList, &ts)
		}

		return nil
	}

	g := multierror.Group{}

	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				if err := handleData(buf[:n]); err != nil {
					return tipsetList{}, err
				}
				break
			}
			return tipsetList{}, err
		}
		extra, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return tipsetList{}, err
		}

		g.Go(func() error {
			c <- struct{}{}
			b := make([]byte, len(buf[:n])+len(extra))
			copy(b, buf[:n])
			err := handleData(append(b, extra...))
			<-c
			return err
		})
	}

	multiErr := g.Wait()
	if multiErr != nil {
		return tsList, err
	}

	return tsList, nil
}
