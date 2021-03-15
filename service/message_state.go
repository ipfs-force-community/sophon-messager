package service

import (
	"sync"
	"time"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

type MessageState struct {
	repo repo.Repo
	log  *logrus.Logger
	cfg  *config.MessageStateConfig

	idCids *idCidCache // 保存 cid 和 id的映射，方便从msgCache中找消息状态

	messageCache *cache.Cache // id 为 key

	l sync.Mutex
}

func NewMessageState(repo repo.Repo, logger *logrus.Logger, cfg *config.MessageStateConfig) (*MessageState, error) {
	ms := &MessageState{
		repo: repo,
		log:  logger,
		cfg:  cfg,
		idCids: &idCidCache{
			cache: make(map[string]types.UUID),
		},
		messageCache: cache.New(time.Duration(cfg.DefaultExpiration)*time.Second, time.Duration(cfg.CleanupInterval)*time.Second),
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
		if msg.UnsignedCid.Defined() {
			ms.idCids.Set(msg.UnsignedCid.String(), msg.ID)
			ms.SetMessage(msg.ID, msg)
		}
	}
	return nil
}

func (ms *MessageState) GetMessage(id types.UUID) (*types.Message, bool) {
	v, ok := ms.messageCache.Get(id.String())
	if ok {
		return v.(*types.Message), ok
	}

	return nil, ok
}

func (ms *MessageState) SetMessage(id types.UUID, message *types.Message) {
	ms.messageCache.SetDefault(id.String(), message)
}

func (ms *MessageState) DeleteMessage(id types.UUID) {
	ms.messageCache.Delete(id.String())
}

func (ms *MessageState) MutatorMessage(id types.UUID, f func(*types.Message) error) error {
	var msg *types.Message
	if v, ok := ms.messageCache.Get(id.String()); ok {
		msg = v.(*types.Message)
	} else {
		var err error
		msg, err = ms.repo.MessageRepo().GetMessage(id)
		if err != nil {
			ms.log.Errorf("get message failed, id: %v, err: %v", id, err)
			return err
		}
	}

	err := f(msg)
	if err != nil {
		return err
	}
	ms.messageCache.SetDefault(id.String(), msg)
	return nil
}

func (ms *MessageState) UpdateMessageStateByCid(cid string, state types.MessageState) error {
	id, ok := ms.idCids.Get(cid)
	if !ok {
		msg, err := ms.repo.MessageRepo().GetMessageByCid(cid)
		if err != nil {
			return err
		}
		ms.SetMessage(msg.ID, msg)
		return nil
	}

	return ms.MutatorMessage(id, func(message *types.Message) error {
		message.State = state
		return nil
	})
}

func (ms *MessageState) GetMessageStateByCid(cid string) (types.MessageState, bool) {
	id, ok := ms.idCids.Get(cid)
	if !ok {
		return types.UnKnown, ok
	}
	msg, ok := ms.GetMessage(id)
	if !ok || msg == nil {
		return types.UnKnown, ok
	}

	return msg.State, ok
}
