package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/big"
	chain2 "github.com/filecoin-project/venus/app/submodule/chain"
	"github.com/filecoin-project/venus/pkg/types"
	"github.com/google/uuid"
	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/service"
	types2 "github.com/ipfs-force-community/venus-messager/types"
	"github.com/ipfs-force-community/venus-messager/utils"
	"github.com/ipfs/go-cid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func objectToString(i interface{}) string {
	res, _ := json.MarshalIndent(i, "", " ")
	return string(res)
}

var db repo.Repo

var venusApi, closer = (*service.NodeClient)(nil), jsonrpc.ClientCloser(nil)

var ctx = context.TODO()
var msgService *service.MessageService

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func shutdown() {
	if db == nil {
		return
	}
	if err := db.DbClose(); err != nil {
		fmt.Printf("shutdown postgre client failed. %v", err)
	}
}

func setup() {
	var err error
	if db, err = OpenSqlite(&config.SqliteConfig{Path: "sqlite.db"}); err != nil {
		panic(err)
	}
	if err = db.AutoMigrate(); err != nil {
		panic(err)
	}

	if venusApi, closer, err = service.NewNodeClient(ctx, &config.NodeConfig{
		Url:   "/ip4/192.168.1.134/tcp/3453",
		Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJhbGwiXX0.n0eSFUWCbosjteqktQOcOghw7VWOm5wODkgpoT2yFJw"}); err != nil {
		panic(err)
	}
	msgService = service.NewMessageService(db, venusApi, logrus.New())
}

func loadTestMessages() ([]*types2.Message, error) {
	var head *types.TipSet
	var err error
	var messages = make(map[cid.Cid]*types.UnsignedMessage)

goOut:
	for len(messages) < 10 {
		if head, err = venusApi.ChainGetTipSet(ctx, head.Parents()); err != nil {
			return nil, err
		}
		blocks := head.Blocks()
		for _, block := range blocks {
			var blockMsgs *chain2.BlockMessages
			if blockMsgs, err = venusApi.ChainGetBlockMessages(ctx, block.Cid()); err != nil {
				return nil, err
			}
			for _, msg := range blockMsgs.BlsMessages {
				messages[msg.Cid()] = msg
				if len(messages) >= 10 {
					break goOut
				}
			}
		}
	}

	var unsignedMsgs = make([]*types2.Message, len(messages))
	var idx = 0
	for _, msg := range messages {
		unsignedMsgs[idx] = &types2.Message{
			Uid:             uuid.New().String(),
			UnsignedMessage: *msg,
			Signature:       nil,
			Epoch:           0,
			Receipt:         nil,
			Meta: &types2.MsgMeta{
				ExpireEpoch:       1024,
				GasOverEstimation: 0,
				MaxFee:            big.NewInt(200),
				MaxFeeCap:         big.Int{},
			},
		}
		idx++
	}
	return unsignedMsgs, nil
}

func TestSaveMessages(t *testing.T) {
	var msgs, err = loadTestMessages()
	assert.NoError(t, err)

	messageRepo := db.MessageRepo()
	for _, msg := range msgs {
		_, err = messageRepo.SaveMessage(msg)
		assert.NoError(t, err)
	}
}

func TestUpdateMessageState(t *testing.T) {
	// TestSaveMessages(t)
	err := msgService.DoRefreshMsgsState()
	assert.NoError(t, err)
}

func TestMessage(t *testing.T) {
	msgDb := db.MessageRepo()
	msg := utils.NewTestMsg()
	beforeSave := objectToString(msg)

	t.Logf("%s", beforeSave)

	uuid, err := msgDb.SaveMessage(msg)
	assert.NoError(t, err)

	result, err := msgDb.GetMessage(uuid)
	assert.NoError(t, err)
	afterSave := objectToString(result)

	assert.Equal(t, beforeSave, afterSave)

	t.Logf("%s", afterSave)

	allMsg, err := msgDb.ListMessage()
	assert.NoError(t, err)
	assert.LessOrEqual(t, 1, len(allMsg))

	unchainedMsgs, err := msgDb.ListUnchainedMsgs()
	assert.NoError(t, err)
	assert.LessOrEqual(t, 1, len(unchainedMsgs))
}
