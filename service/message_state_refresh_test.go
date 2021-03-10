package service

import (
	"context"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/filecoin-project/go-jsonrpc"
	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/models/sqlite"
	"github.com/ipfs-force-community/venus-messager/types"
)

type builder struct {
	ctx              context.Context
	repo             repo.Repo
	venusClient      *NodeClient
	venusClientClose jsonrpc.ClientCloser
	msgService       *MessageService
	addrService      *AddressService
	walletService    *WalletService
	event            *NodeEvents
}

func build(t *testing.T) *builder {
	db, err := sqlite.OpenSqlite(&config.SqliteConfig{Path: "sqlite.db"})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate())

	venusApi, closer, err := NewNodeClient(context.Background(), &config.NodeConfig{
		Url:   "/ip4/192.168.1.134/tcp/3453",
		Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJhbGwiXX0.n0eSFUWCbosjteqktQOcOghw7VWOm5wODkgpoT2yFJw"})
	assert.NoError(t, err)

	log := logrus.New()

	walletService, err := NewWalletService(db, log)
	assert.NoError(t, err)

	addressService, err := NewAddressService(db, log, walletService, venusApi, &config.AddressConfig{RemoteWalletSweepInterval: 10})
	assert.NoError(t, err)

	messageServiceCfg := &config.MessageServiceConfig{"tipset.db"}
	cfg := &config.Config{MessageService: *messageServiceCfg}
	config.CheckFile(cfg)

	msgService, err := NewMessageService(db, venusApi, log, messageServiceCfg,
		NewMessageState(db, logrus.New(), &config.MessageStateConfig{BackTime: 3600}), addressService)
	assert.NoError(t, err)

	event := &NodeEvents{
		client:     *venusApi,
		log:        log,
		msgService: msgService,
	}

	return &builder{
		ctx:              context.TODO(),
		repo:             db,
		venusClient:      venusApi,
		venusClientClose: closer,
		msgService:       msgService,
		walletService:    walletService,
		addrService:      addressService,
		event:            event,
	}
}

func (b *builder) LoadMessage(count int) ([]*types.Message, error) {
	var ts *venustypes.TipSet
	msgs := make([]*types.Message, 0, count)
	for len(msgs) <= count {
		ts, err := b.venusClient.ChainGetTipSet(b.ctx, ts.Parents())
		if err != nil {
			return nil, err
		}
		for _, block := range ts.Blocks() {
			blockMsgs, err := b.venusClient.ChainGetBlockMessages(b.ctx, block.Cid())
			if err != nil {
				return nil, err
			}
			for _, m := range blockMsgs.SecpkMessages {
				msgs = append(msgs, &types.Message{
					ID:              uuid.New().String(),
					UnsignedMessage: m.Message,
					Signature:       &m.Signature,
					State:           types.Unsigned,
				})
			}
			for _, m := range blockMsgs.BlsMessages {
				msgs = append(msgs, &types.Message{
					ID:              uuid.New().String(),
					UnsignedMessage: *m,
					State:           types.Unsigned,
				})
			}
		}
	}

	return msgs, nil
}

func TestMessageStateRefresh(t *testing.T) {
	builder := build(t)
	msgs, err := builder.LoadMessage(10)
	assert.NoError(t, err)
	for _, m := range msgs {
		_, err := builder.repo.MessageRepo().SaveMessage(m)
		assert.NoError(t, err)
	}

	builder.event.listenHeadChangesOnce(builder.ctx)

	time.Sleep(time.Hour * 3)
}

func TestReadAndWriteTipset(t *testing.T) {
	var tsList tipsetList
	tsList = append(tsList, &tipsetFormat{
		Cid:    []string{"00000"},
		Height: 0,
	})
	tsList = append(tsList, &tipsetFormat{
		Cid:    []string{"33333"},
		Height: 3,
	})
	tsList = append(tsList, &tipsetFormat{
		Cid:    []string{"22222"},
		Height: 2,
	})

	filePath := "./test_read_write_tipset.db"
	defer func() {
		assert.NoError(t, os.Remove(filePath))
	}()
	err := updateTipsetFile(filePath, tsList)
	assert.NoError(t, err)

	result, err := readTipsetFromFile(filePath)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	t.Logf("before sort %+v", result)

	sort.Sort(result)
	t.Logf("after sort %+v", result)
	assert.Equal(t, tsList[1], result[0])
}

func TestResetTipsetFile(t *testing.T) {
	path := "./test.db"
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 06666)
	assert.NoError(t, err)

	f.WriteString("ddddddd")
	f.Close()

	assert.NoError(t, resetTipsetFile(path))
}
