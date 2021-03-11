package service

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type MessageState struct {
	repo repo.Repo
	log  *logrus.Logger
	cfg  *config.MessageStateConfig

	idCids *idCidCache // 保存 cid 和 id的映射，方便从msgCache中找消息状态

	msgState map[string]types.MessageState // id 为 key

	l sync.Mutex
}

func NewMessageState(repo repo.Repo, logger *logrus.Logger, cfg *config.MessageStateConfig) (*MessageState, error) {
	ms := &MessageState{
		repo: repo,
		log:  logger,
		cfg:  cfg,
		idCids: &idCidCache{
			cache: make(map[string]string),
		},
		msgState: make(map[string]types.MessageState),
	}

	if err := ms.loadRecentMessage(); err != nil {
		return nil, err
	}

	return ms, nil
}

func (ms *MessageState) loadRecentMessage() error {
	startTime := time.Now().Add(-time.Second * time.Duration(ms.cfg.BackTime))
	msgs, err := ms.repo.MessageRepo().GetSignedMessageByTime(startTime)
	if err != nil {
		return err
	}
	ms.log.Infof("load recent message: %d", len(msgs))

	for _, msg := range msgs {
		if msg.UnsignedCid != nil {
			ms.idCids.Set(msg.ID.String(), msg.UnsignedCid.String())
			ms.SetMessageState(msg.ID.String(), msg.State)
		}
	}
	return nil
}

func (ms *MessageState) GetMessageState(id string) (types.MessageState, bool) {
	ms.l.Lock()
	defer ms.l.Unlock()
	v, ok := ms.msgState[id]

	return v, ok
}

func (ms *MessageState) SetMessageState(id string, state types.MessageState) {
	ms.l.Lock()
	defer ms.l.Unlock()

	ms.msgState[id] = state
}

func (ms *MessageState) DeleteMessageState(id string) {
	ms.l.Lock()
	defer ms.l.Unlock()

	delete(ms.msgState, id)
}

func (ms *MessageState) UpdateMessageStateByCid(cid string, state types.MessageState) error {
	id, ok := ms.idCids.Get(cid)
	if !ok {
		msg, err := ms.repo.MessageRepo().GetMessageByCid(cid)
		if err != nil {
			return err
		}
		ms.SetMessageState(msg.ID.String(), state)
		return nil
	}

	ms.SetMessageState(id, state)
	return nil
}

type idCidCache struct {
	cache map[string]string
	l     sync.Mutex
}

func (ic *idCidCache) Set(cid, id string) {
	ic.l.Lock()
	defer ic.l.Unlock()
	ic.cache[cid] = id
}

func (ic *idCidCache) Get(cid string) (string, bool) {
	ic.l.Lock()
	defer ic.l.Unlock()
	cid, ok := ic.cache[cid]

	return cid, ok
}
