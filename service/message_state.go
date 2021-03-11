package service

import (
	"sync"
	"time"

	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type MessageState struct {
	repo repo.Repo
	log  *logrus.Logger
	cfg  *config.MessageStateConfig

	messageCache *cache.Cache
	idCids       *idCidCache // 保存 cid 和 id的映射，方便从messageCache中找消息

	l sync.Mutex
}

func NewMessageState(repo repo.Repo, logger *logrus.Logger, cfg *config.MessageStateConfig) *MessageState {
	return &MessageState{
		repo:         repo,
		log:          logger,
		cfg:          cfg,
		messageCache: cache.New(time.Duration(cfg.DefaultExpiration)*time.Second, time.Duration(cfg.CleanupInterval)*time.Second),
		idCids: &idCidCache{
			cache: make(map[string]string),
		},
	}
}

func (ms *MessageState) loadRecentMessage() error {
	startTime := time.Now().Add(-time.Second * time.Duration(ms.cfg.BackTime))
	msgs, err := ms.repo.MessageRepo().GetMessageByTime(startTime)
	if err != nil {
		return err
	}
	ms.log.Infof("load recent message: %d", len(msgs))
	ms.SetMessages(msgs)

	for _, msg := range msgs {
		if msg.SignedCid().Defined() {
			ms.idCids.Set(msg.ID, msg.SignedCid().String())
		}
	}
	return nil
}

func (ms *MessageState) GetMessage(id string) (*types.Message, bool) {
	v, ok := ms.messageCache.Get(id)
	if ok {
		return v.(*types.Message), ok
	}

	return nil, ok
}

func (ms *MessageState) SetMessage(msg *types.Message) {
	ms.messageCache.SetDefault(msg.ID, msg)
}

func (ms *MessageState) SetMessages(msgs []*types.Message) {
	for _, msg := range msgs {
		ms.SetMessage(msg)
	}
}

func (ms *MessageState) DeleteMessage(id string) {
	ms.messageCache.Delete(id)
}

func (ms *MessageState) UpdateMessageState(id string, state types.MessageState) {
	if v, ok := ms.messageCache.Get(id); ok {
		msg := v.(*types.Message)
		msg.State = state
		ms.messageCache.SetDefault(id, msg)
	} else {
		m, err := ms.repo.MessageRepo().GetMessage(id)
		if err != nil {
			ms.log.Errorf("get message failed, id: %v, err: %v", id, err)
			return
		}
		m.State = state
		ms.messageCache.SetDefault(id, m)
	}
}

func (ms *MessageState) UpdateMessageStateAndReceipt(cidStr string, state types.MessageState, receipt *venustypes.MessageReceipt) {
	if id, ok := ms.idCids.Get(cidStr); ok {
		if m, ok := ms.GetMessage(id); ok {
			m.State = state
			if receipt != nil {
				m.Receipt = receipt
			}
		}
	} else {
		m, err := ms.repo.MessageRepo().GetMessageByCid(cidStr)
		if err != nil {
			ms.log.Errorf("get message by cid failed, cid: %v, err: %v", cidStr, err)
			return
		}
		m.State = state
		if receipt != nil {
			m.Receipt = receipt
		}
		ms.SetMessage(m)
		ms.idCids.Set(m.ID, cidStr)
	}
}

type idCidCache struct {
	cache map[string]string
	l     sync.Mutex
}

func (ic *idCidCache) Set(id, cid string) {
	ic.l.Lock()
	defer ic.l.Unlock()
	ic.cache[cid] = id
}

func (ic *idCidCache) Get(cid string) (string, bool) {
	ic.l.Lock()
	defer ic.l.Unlock()

	id, ok := ic.cache[cid]
	return id, ok
}
