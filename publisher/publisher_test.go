package publisher

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/filecoin-project/venus-messager/testhelper"
	mockV1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1/mock"
	"github.com/filecoin-project/venus/venus-shared/types"
	mtypes "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMainNodePublishMessage(t *testing.T) {
	ctx := context.Background()
	// mock api
	ctrl := gomock.NewController(t)
	mainNode := mockV1.NewMockFullNode(ctrl)

	rpcPublisher := NewRpcPublisher(ctx, mainNode, nil, false)
	publisher := NewMergePublisher(ctx, rpcPublisher)
	msgs := testhelper.NewShareSignedMessages(10)

	mainNode.EXPECT().MpoolBatchPush(ctx, msgs).Return(nil, nil).Times(1)
	err := publisher.PublishMessages(ctx, msgs)
	assert.NoError(t, err)
	runtime.Gosched()
	time.Sleep(1 * time.Second)
}

func TestMultiNodePublishMessage(t *testing.T) {
	ctx := context.Background()
	msgs := testhelper.NewShareSignedMessages(5)

	// mock api
	ctrl := gomock.NewController(t)
	mainNode := mockV1.NewMockFullNode(ctrl)
	mainNode.EXPECT().MpoolBatchPush(ctx, msgs).Return(nil, nil).AnyTimes()

	servers := make([]*testhelper.FullNodeServer, 4)
	for i := 0; i < 4; i++ {
		e, err := testhelper.MockFullNodeServer(t)
		assert.NoError(t, err)
		servers[i] = e
	}
	nodes := make([]*mtypes.Node, 4)
	for i := 0; i < 4; i++ {
		nodes[i] = &mtypes.Node{
			ID:    types.NewUUID(),
			Name:  "node_" + strconv.Itoa(i),
			URL:   fmt.Sprintf("/ip4/127.0.0.1/tcp/%s", servers[i].Port),
			Token: servers[i].Token,
			Type:  mtypes.NodeType(rand.Intn(3)),
		}
	}

	nodeProvider := testhelper.NewMockNodeRepo(ctrl)
	rpcPublisher := NewRpcPublisher(ctx, mainNode, nodeProvider, true)

	t.Run("publish message to multi node", func(t *testing.T) {
		nodeProvider.EXPECT().ListNode().Return(nodes[:3], nil).Times(1)
		for _, srv := range servers[:3] {
			srv.FullNode.EXPECT().MpoolBatchPush(gomock.Any(), msgs).Return(nil, nil).Times(1)
		}
		err := rpcPublisher.PublishMessages(ctx, msgs)
		assert.NoError(t, err)
		runtime.Gosched()
	})

	// wait for messager consume
	time.Sleep(1 * time.Second)

	t.Run("publish message to multi node after delete node", func(t *testing.T) {
		nodeProvider.EXPECT().ListNode().Return(nodes[1:2], nil).Times(1)
		for _, srv := range servers[1:2] {
			srv.FullNode.EXPECT().MpoolBatchPush(gomock.Any(), msgs).Return(nil, nil).Times(1)
		}
		err := rpcPublisher.PublishMessages(ctx, msgs)
		assert.NoError(t, err)
		runtime.Gosched()
	})

	t.Run("publish message to multi node after add node", func(t *testing.T) {
		nodeProvider.EXPECT().ListNode().Return(nodes[:4], nil).Times(1)
		for _, srv := range servers[:4] {
			srv.FullNode.EXPECT().MpoolBatchPush(gomock.Any(), msgs).Return(nil, nil).Times(1)
		}
		err := rpcPublisher.PublishMessages(ctx, msgs)
		assert.NoError(t, err)
		runtime.Gosched()
	})

	// wait goroutine
	time.Sleep(1 * time.Second)
}

func TestMergePublisher(t *testing.T) {
	ctx := context.Background()
	// mock api
	ctrl := gomock.NewController(t)
	p1 := testhelper.NewMockIMsgPublisher(ctrl)
	p2 := testhelper.NewMockIMsgPublisher(ctrl)

	publisher := NewMergePublisher(ctx, p1, p2)
	msgs := testhelper.NewShareSignedMessages(10)

	p1.EXPECT().PublishMessages(ctx, msgs).Return(nil).Times(1)
	p2.EXPECT().PublishMessages(ctx, msgs).Return(nil).Times(1)

	err := publisher.PublishMessages(ctx, msgs)
	assert.NoError(t, err)
}

func TestMsgCache(t *testing.T) {
	ctx := context.Background()
	// mock api
	ctrl := gomock.NewController(t)
	iPublisher := testhelper.NewMockIMsgPublisher(ctrl)

	publisher, err := NewCachePublisher(ctx, 1, iPublisher)
	assert.NoError(t, err)
	msgs := testhelper.NewShareSignedMessages(10)

	iPublisher.EXPECT().PublishMessages(ctx, msgs[:4]).Return(nil).Times(1)
	err = publisher.PublishMessages(ctx, msgs[:4])
	assert.NoError(t, err)
	runtime.Gosched()

	iPublisher.EXPECT().PublishMessages(ctx, msgs[4:]).Return(nil).Times(1)
	err = publisher.PublishMessages(ctx, msgs)
	assert.NoError(t, err)
	runtime.Gosched()

	err = publisher.PublishMessages(ctx, msgs)
	assert.NoError(t, err)
	runtime.Gosched()

	// wait cache to be expired
	time.Sleep(3 * time.Second)
	iPublisher.EXPECT().PublishMessages(ctx, msgs).Return(nil).Times(1)
	err = publisher.PublishMessages(ctx, msgs)
	assert.NoError(t, err)
	runtime.Gosched()
	time.Sleep(1 * time.Second)
}

func TestConcurrentPublisher(t *testing.T) {
	ctx := context.Background()
	// mock api
	ctrl := gomock.NewController(t)
	iPublisher := testhelper.NewMockIMsgPublisher(ctrl)

	publisher, err := NewConcurrentPublisher(ctx, 2, iPublisher)
	assert.NoError(t, err)
	msgs := testhelper.NewShareSignedMessages(10)

	iPublisher.EXPECT().PublishMessages(ctx, msgs).Return(nil).Times(1)
	err = publisher.PublishMessages(ctx, msgs)
	assert.NoError(t, err)
	runtime.Gosched()

	time.Sleep(1 * time.Second)
}

func TestIntergrate(t *testing.T) {
	ctx := context.Background()
	// mock api
	ctrl := gomock.NewController(t)
	p1 := testhelper.NewMockIMsgPublisher(ctrl)
	p2 := testhelper.NewMockIMsgPublisher(ctrl)

	mergePublisher := NewMergePublisher(ctx, p1, p2)
	concurrentPublisher, err := NewConcurrentPublisher(ctx, 2, mergePublisher)
	assert.NoError(t, err)
	cachePublisher, err := NewCachePublisher(ctx, 5, concurrentPublisher)
	assert.NoError(t, err)

	msgs := testhelper.NewShareSignedMessages(10)
	p1.EXPECT().PublishMessages(ctx, msgs).Return(nil).Times(1)
	p2.EXPECT().PublishMessages(ctx, msgs).Return(nil).Times(1)

	err = cachePublisher.PublishMessages(ctx, msgs)
	assert.NoError(t, err)
	runtime.Gosched()
	time.Sleep(1 * time.Second)
}
