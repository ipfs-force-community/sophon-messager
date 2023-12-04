package publisher

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/ipfs-force-community/sophon-messager/filestore"
	"github.com/ipfs-force-community/sophon-messager/mocks"
	"github.com/ipfs-force-community/sophon-messager/models/sqlite"
	"github.com/ipfs-force-community/sophon-messager/testhelper"

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

	fs := filestore.NewMockFileStore(t.TempDir())
	sqliteRepo, err := sqlite.OpenSqlite(fs)
	assert.NoError(t, err)
	assert.NoError(t, sqliteRepo.AutoMigrate())

	rpcPublisher := NewRpcPublisher(ctx, mainNode, nil, false, sqliteRepo.MessageRepo())
	publisher := NewMergePublisher(ctx, rpcPublisher)
	msgs := testhelper.NewShareSignedMessages(10)

	mainNode.EXPECT().MpoolBatchPushUntrusted(ctx, msgs).Return(nil, nil).Times(1)
	err = publisher.PublishMessages(ctx, msgs)
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
	mainNode.EXPECT().MpoolBatchPushUntrusted(ctx, msgs).Return(nil, nil).AnyTimes()

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

	fs := filestore.NewMockFileStore(t.TempDir())
	sqliteRepo, err := sqlite.OpenSqlite(fs)
	assert.NoError(t, err)
	assert.NoError(t, sqliteRepo.AutoMigrate())
	nodeProvider := mocks.NewMockNodeRepo(ctrl)
	rpcPublisher := NewRpcPublisher(ctx, mainNode, nodeProvider, true, sqliteRepo.MessageRepo())

	t.Run("publish message to multi node", func(t *testing.T) {
		nodeProvider.EXPECT().ListNode().Return(nodes[:3], nil).Times(1)
		for _, srv := range servers[:3] {
			srv.FullNode.EXPECT().MpoolBatchPushUntrusted(gomock.Any(), msgs).Return(nil, nil).Times(1)
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
			srv.FullNode.EXPECT().MpoolBatchPushUntrusted(gomock.Any(), msgs).Return(nil, nil).Times(1)
		}
		err := rpcPublisher.PublishMessages(ctx, msgs)
		assert.NoError(t, err)
		runtime.Gosched()
	})

	t.Run("publish message to multi node after add node", func(t *testing.T) {
		nodeProvider.EXPECT().ListNode().Return(nodes[:4], nil).Times(1)
		for _, srv := range servers[:4] {
			srv.FullNode.EXPECT().MpoolBatchPushUntrusted(gomock.Any(), msgs).Return(nil, nil).Times(1)
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
	p1 := mocks.NewMockIMsgPublisher(ctrl)
	p2 := mocks.NewMockIMsgPublisher(ctrl)

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
	iPublisher := mocks.NewMockIMsgPublisher(ctrl)

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
	iPublisher := mocks.NewMockIMsgPublisher(ctrl)

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
	p1 := mocks.NewMockIMsgPublisher(ctrl)
	p2 := mocks.NewMockIMsgPublisher(ctrl)

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

func TestPublishMessageFailed(t *testing.T) {
	ctx := context.Background()
	// mock api
	ctrl := gomock.NewController(t)
	mainNode := mockV1.NewMockFullNode(ctrl)

	fs := filestore.NewMockFileStore(t.TempDir())
	sqliteRepo, err := sqlite.OpenSqlite(fs)
	assert.NoError(t, err)
	assert.NoError(t, sqliteRepo.AutoMigrate())

	rpcPublisher := NewRpcPublisher(ctx, mainNode, nil, false, sqliteRepo.MessageRepo())
	publisher := NewMergePublisher(ctx, rpcPublisher)

	msgs := testhelper.NewShareSignedMessages(10)
	form := msgs[0].Message.From
	for _, msg := range msgs {
		msgCid := msg.Cid()
		msg.Message.From = form
		m := &mtypes.Message{
			ID:          types.NewUUID().String(),
			UnsignedCid: &msgCid,
			SignedCid:   &msgCid,
			Message:     msg.Message,
		}
		assert.NoError(t, sqliteRepo.MessageRepo().CreateMessage(m))
	}

	balanceToLowErr := fmt.Errorf("not enough funds (required: 0.08343657656301909 FIL, balance: 0.003413734154635385 FIL): not enough funds to execute transaction")
	mainNode.EXPECT().MpoolBatchPushUntrusted(ctx, msgs).Return(nil, balanceToLowErr).Times(1)
	err = publisher.PublishMessages(ctx, msgs)
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		msg, err := sqliteRepo.MessageRepo().GetMessageByCid(msgs[0].Cid())
		if err == nil && len(msg.ErrorMsg) != 0 {
			assert.Equal(t, balanceToLowErr.Error(), msg.ErrorMsg)
			break
		}

		if i == 9 {
			assert.Fail(t, "failed to get message error")
		}

		time.Sleep(1 * time.Second)
	}
}
