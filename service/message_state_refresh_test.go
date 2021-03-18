package service

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/filecoin-project/go-jsonrpc"
	venustypes "github.com/filecoin-project/venus/pkg/types"
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

	messageServiceCfg := &config.MessageServiceConfig{TipsetFilePath: "tipset.txt", IsProcessHead: true}
	cfg := &config.Config{MessageService: *messageServiceCfg}
	assert.NoError(t, config.CheckFile(cfg))

	messageState, err := NewMessageState(db, logrus.New(), &config.MessageStateConfig{BackTime: 3600})
	assert.NoError(t, err)

	msgService, err := NewMessageService(db, venusApi, log, messageServiceCfg, messageState, addressService)
	assert.NoError(t, err)

	event := &NodeEvents{
		client:     venusApi,
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
					ID:              types.NewUUID(),
					UnsignedMessage: m.Message,
					Signature:       &m.Signature,
					State:           types.UnFillMsg,
				})
			}
			for _, m := range blockMsgs.BlsMessages {
				msgs = append(msgs, &types.Message{
					ID:              types.NewUUID(),
					UnsignedMessage: *m,
					State:           types.UnFillMsg,
				})
			}
		}
	}

	return msgs, nil
}

func TestMessageStateRefresh(t *testing.T) {
	t.Skip()
	builder := build(t)
	msgs, err := builder.LoadMessage(10)
	assert.NoError(t, err)
	for _, m := range msgs {
		_, err := builder.repo.MessageRepo().SaveMessage(m)
		assert.NoError(t, err)
	}

	assert.NoError(t, builder.event.listenHeadChangesOnce(builder.ctx))

	time.Sleep(time.Minute * 6)
}

func TestReadAndWriteTipset(t *testing.T) {
	tsCache := &TipsetCache{Cache: map[int64]*tipsetFormat{}, CurrHeight: 0}
	tsCache.Cache[0] = &tipsetFormat{
		Key:    "00000",
		Height: 0,
	}
	tsCache.Cache[3] = &tipsetFormat{
		Key:    "33333",
		Height: 3,
	}
	tsCache.Cache[2] = &tipsetFormat{
		Key:    "22222",
		Height: 2,
	}
	tsCache.CurrHeight = 3

	filePath := "./test_read_write_tipset.txt"
	defer func() {
		//assert.NoError(t, os.Remove(filePath))
	}()
	err := updateTipsetFile(filePath, tsCache)
	assert.NoError(t, err)

	result, err := readTipsetFile(filePath)
	assert.NoError(t, err)
	assert.Len(t, result.Cache, 3)

	var tsList tipsetList
	for _, c := range result.Cache {
		tsList = append(tsList, c)
	}
	t.Logf("before sort %+v", tsList)

	sort.Sort(tsList)
	t.Logf("after sort %+v", tsList)
	assert.Equal(t, tsList[1].Height, int64(2))
}
