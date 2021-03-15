package service

import (
	"context"
	"testing"

	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/venus-messager/config"
)

func TestNewAddressClient(t *testing.T) {
	t.Skip()
	// a valid URL and token are required
	url := "/ip4/0.0.0.0/tcp/5678"
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.GuKxM-lRDRdbSUwlhERzsF8hJK14XEcFYgWdlICHM4I"
	cli, close, err := newWalletClient(context.Background(), url, token)
	assert.NoError(t, err)
	defer close()

	addrs, err := cli.WalletList(context.Background())
	assert.NoError(t, err)
	t.Log("address: ", addrs)
	assert.Equal(t, 3, len(addrs))

	token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJhbGwiXX0.U98efTo3RgQjL39L1_1d4xgHWi_ttqaMbAMczorV0Ww"
	url = "/ip4/127.0.0.1/tcp/3453"

	cfg := &config.NodeConfig{
		Url:   url,
		Token: token,
	}
	node, close, err := NewNodeClient(context.Background(), cfg)
	assert.NoError(t, err)
	defer close()
	for _, addr := range addrs {
		actor, err := node.StateGetActor(context.Background(), addr, venustypes.EmptyTSK)
		t.Log(actor, err)
	}
}
