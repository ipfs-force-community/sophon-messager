package publisher

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/filecoin-project/venus-messager/models/repo"
	mpubsub "github.com/filecoin-project/venus-messager/publisher/pubsub"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs/go-cid"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

var errAlreadyInMpool = fmt.Errorf("already in mpool: validation failure")
var errMinimumNonce = errors.New("minimum expected nonce")

//go:generate mockgen -destination=../mocks/mock_msg_publisher.go -package=mocks github.com/filecoin-project/venus-messager/publisher IMsgPublisher

type IMsgPublisher interface {
	// PublishMessages publish messages to chain
	PublishMessages(ctx context.Context, msgs []*types.SignedMessage) error
}

type P2pPublisher struct {
	topic *pubsub.Topic
}

func NewP2pPublisher(pubsub mpubsub.IPubsuber, netName types.NetworkName) (*P2pPublisher, error) {
	topicName := fmt.Sprintf("/fil/msgs/%s", netName)
	topic, err := pubsub.GetTopic(topicName)
	if err != nil {
		return nil, err
	}

	return &P2pPublisher{
		topic: topic,
	}, nil
}

func (p *P2pPublisher) PublishMessages(ctx context.Context, msgs []*types.SignedMessage) error {
	for _, msg := range msgs {
		msgb, err := msg.Serialize()
		if err != nil {
			return fmt.Errorf("marshal message %s failed %w", msg.Cid(), err)
		}
		if err := p.topic.Publish(ctx, msgb); err != nil {
			return fmt.Errorf("publish message %s failed %w", msg.Cid(), err)
		}
	}
	return nil
}

type RpcPublisher struct {
	ctx             context.Context
	mainNodeThread  *nodeThread
	nodeProvider    repo.INodeProvider
	enableMultiNode bool

	nodeThreads map[types.UUID]struct {
		nodeThread *nodeThread
		close      func()
	}
	lk sync.Mutex
}

func NewRpcPublisher(ctx context.Context, nodeClient v1.FullNode, nodeProvider repo.INodeProvider, enableMultiNode bool) *RpcPublisher {
	nThread := newNodeThread(ctx, "mainNode", nodeClient)
	return &RpcPublisher{
		ctx:             ctx,
		mainNodeThread:  nThread,
		nodeProvider:    nodeProvider,
		enableMultiNode: enableMultiNode,
		nodeThreads: make(map[types.UUID]struct {
			nodeThread *nodeThread
			close      func()
		}),

		lk: sync.Mutex{},
	}
}

func (p *RpcPublisher) PublishMessages(ctx context.Context, msgs []*types.SignedMessage) error {
	p.mainNodeThread.HandleMsg(msgs)

	if !p.enableMultiNode {
		return nil
	}

	nodeList, err := p.nodeProvider.ListNode()
	if err != nil {
		return fmt.Errorf("list node fail %w", err)
	}

	p.lk.Lock()
	defer p.lk.Unlock()

	nodesRemain := make(map[types.UUID]struct{})
	for _, node := range nodeList {
		threadStruct, ok := p.nodeThreads[node.ID]
		nodesRemain[node.ID] = struct{}{}
		if !ok {
			thrCtx, cancel := context.WithCancel(p.ctx) // nolint ignore lostcancel
			cli, closer, err := v1.DialFullNodeRPC(thrCtx, node.URL, node.Token, nil)
			if err != nil {
				log.Warnf("connect node(%s) fail %v", node.Name, err)
				continue
			}

			nodeName := node.Name
			threadStruct = struct {
				nodeThread *nodeThread
				close      func()
			}{
				nodeThread: newNodeThread(thrCtx, nodeName, cli),
				close: func() {
					cancel()
					closer()
					log.Debugf("close node thread %s", nodeName)
				},
			}
			p.nodeThreads[node.ID] = threadStruct
		}
		threadStruct.nodeThread.HandleMsg(msgs)
	}

	for id, threadStruct := range p.nodeThreads {
		if _, ok := nodesRemain[id]; !ok {
			threadStruct.close()
			delete(p.nodeThreads, id)
		}
	}

	return nil // nolint ignore lostcancel
}

type nodeThread struct {
	name       string
	nodeClient v1.FullNode
	msgChan    chan []*types.SignedMessage
}

func newNodeThread(ctx context.Context, name string, nodeClient v1.FullNode) *nodeThread {
	t := &nodeThread{
		name:       name,
		nodeClient: nodeClient,
		msgChan:    make(chan []*types.SignedMessage, 30),
	}
	go t.run(ctx)
	return t
}

func (n *nodeThread) run(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msgs := <-n.msgChan:
				if _, err := n.nodeClient.MpoolBatchPushUntrusted(ctx, msgs); err != nil {
					//skip error
					if !strings.Contains(err.Error(), errMinimumNonce.Error()) && !strings.Contains(err.Error(), errAlreadyInMpool.Error()) {
						log.Errorf("failed to push message node: %s, address: %v, error: %v", n.name, msgs[0].Message.From, err)
					} else {
						log.Debugf("failed to push message %v", err)
					}
				}
			}
		}
	}()
}

func (n *nodeThread) HandleMsg(msgs []*types.SignedMessage) {
	n.msgChan <- msgs
}

type MergePublisher struct {
	ctx           context.Context
	subPublishers []IMsgPublisher
}

func NewMergePublisher(ctx context.Context, publishers ...IMsgPublisher) *MergePublisher {
	m := &MergePublisher{
		ctx:           ctx,
		subPublishers: publishers,
	}
	return m
}

func (p *MergePublisher) PublishMessages(ctx context.Context, msgs []*types.SignedMessage) error {
	if len(p.subPublishers) == 0 {
		return fmt.Errorf("no publisher available")
	}
	for _, publisher := range p.subPublishers {
		err := publisher.PublishMessages(ctx, msgs)
		if err != nil {
			log.Errorf("MergePublisher publish message with sub publisher failed: %v", err)
		}
	}
	return nil
}

func (p *MergePublisher) AddPublisher(publisher IMsgPublisher) {
	p.subPublishers = append(p.subPublishers, publisher)
}

type CachePublisher struct {
	msgCh chan []*types.SignedMessage
	cache map[cid.Cid]bool
	// cacheReleasePeriod is the period of cache release
	cacheReleasePeriod uint64 //seconds
	subPublisher       IMsgPublisher
}

func NewCachePublisher(ctx context.Context, cacheReleasePeriod uint64, subPublisher IMsgPublisher) (*CachePublisher, error) {
	if cacheReleasePeriod == 0 {
		return nil, fmt.Errorf("cache release period should not be zero")
	}
	p := &CachePublisher{
		msgCh:              make(chan []*types.SignedMessage, 30),
		cache:              make(map[cid.Cid]bool),
		cacheReleasePeriod: cacheReleasePeriod,
		subPublisher:       subPublisher,
	}
	p.run(ctx)
	return p, nil
}

func (p *CachePublisher) PublishMessages(ctx context.Context, msgs []*types.SignedMessage) error {
	p.msgCh <- msgs
	return nil
}

func (p *CachePublisher) run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(p.cacheReleasePeriod) * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msgs := <-p.msgCh:
				newMsgs := make([]*types.SignedMessage, 0, len(msgs)/2)
				for _, msg := range msgs {
					c := msg.Cid()
					if _, ok := p.cache[c]; !ok {
						newMsgs = append(newMsgs, msg)
					}
					p.cache[c] = true
				}
				if len(newMsgs) > 0 {
					if err := p.subPublisher.PublishMessages(ctx, newMsgs); err != nil {
						log.Errorf("CachePublisher publish message with sub publisher fail %v", err)
					}
				}
			case <-ticker.C:
				// every cacheReleasePeriod rm old cache ,set new cache to old one
				for k, v := range p.cache {
					if v {
						p.cache[k] = false
					} else {
						delete(p.cache, k)
					}
				}
			}
		}
	}()
}

// ConcurrentPublisher call subPublisher concurrently
type ConcurrentPublisher struct {
	ctx          context.Context
	msgCh        chan []*types.SignedMessage
	subPublisher IMsgPublisher
	concurrency  uint
}

// NewConcurrentPublisher return a ConcurrentPublisher
// subPublisher should be thread safe
func NewConcurrentPublisher(ctx context.Context, concurrency uint, subPublisher IMsgPublisher) (*ConcurrentPublisher, error) {
	if subPublisher == nil {
		return nil, fmt.Errorf("sub publisher is nil")
	}
	c := &ConcurrentPublisher{
		ctx:          ctx,
		msgCh:        make(chan []*types.SignedMessage, 30),
		subPublisher: subPublisher,
		concurrency:  concurrency,
	}
	c.run()
	return c, nil
}

func (p *ConcurrentPublisher) PublishMessages(ctx context.Context, msgs []*types.SignedMessage) error {
	p.msgCh <- msgs
	return nil
}

func (p *ConcurrentPublisher) run() {
	var i uint
	for i = 0; i < p.concurrency; i++ {
		go func() {
			for {
				select {
				case <-p.ctx.Done():
					return
				case msgs := <-p.msgCh:
					err := p.subPublisher.PublishMessages(p.ctx, msgs)
					if err != nil {
						log.Errorf("ConcurrentPublisher publish message with sub publisher fail %v", err)
					}
				}
			}
		}()
	}
}
