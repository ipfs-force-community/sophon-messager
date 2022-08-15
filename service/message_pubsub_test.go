package service

import (
	"bytes"
	"context"
	"testing"
	"time"

	addr "github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/log"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

func TestMessagePubSub(t *testing.T) {
	ctx := context.Background()
	ps1, err := NewMessagePubSub(log.New(), "test_net_name", &config.BootstrapConfig{})
	assert.Nil(t, err)
	addressInfo1 := peer.AddrInfo{
		ID:    ps1.host.ID(),
		Addrs: ps1.host.Addrs(),
	}

	address1, err := peer.AddrInfoToP2pAddrs(&addressInfo1)
	assert.Nil(t, err)

	multiaddr := make([]string, len(address1))
	for i, addr := range address1 {
		multiaddr[i] = addr.String()
	}

	ps2, err := NewMessagePubSub(log.New(), "test_net_name", &config.BootstrapConfig{Addresses: multiaddr})
	assert.Nil(t, err)

	sub, err := ps2.topic.Subscribe()
	assert.Nil(t, err)

	// check conection between ps1 and ps2
	time.Sleep(time.Second * 1)

	assert.Equal(t, 1, len(ps1.host.Network().Peers()))
	assert.Equal(t, 1, len(ps2.host.Network().Peers()))

	// publish message
	msg := types.SignedMessage{
		Message: types.Message{
			From:  addr.TestAddress,
			To:    addr.TestAddress2,
			Value: types.NewInt(100),
		},
	}
	buf := new(bytes.Buffer)
	err = msg.MarshalCBOR(buf)
	assert.Nil(t, err)

	err = ps1.Publish(ctx, &msg)
	assert.Nil(t, err)

	resp, err := sub.Next(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, true, bytes.Equal(resp.Data, buf.Bytes()))

	pi1, err := ps2.FindPeer(ctx, ps1.host.ID())
	assert.Nil(t, err)
	assert.Equal(t, ps1.host.ID(), pi1.ID)

	err = ps2.Connect(ctx, pi1)
	assert.Nil(t, err)
}
