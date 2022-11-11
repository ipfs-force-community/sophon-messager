package pubsub

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
)

func TestMessagePubSub(t *testing.T) {
	ctx := context.Background()
	ps1, err := NewPubsub(ctx, "/ip4/127.0.0.1/tcp/0", "test_net_name", []string{})
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

	ps2, err := NewPubsub(ctx, "/ip4/127.0.0.1/tcp/0", "test_net_name", multiaddr)
	assert.Nil(t, err)

	topic, err := ps1.GetTopic("test")
	assert.Nil(t, err)
	assert.NotNil(t, topic)

	// check connection between ps1 and ps2
	waitTime := 100 // 10s
	for {
		if len(ps1.host.Network().Conns()) > 0 || waitTime <= 0 {
			break
		}
		time.Sleep(time.Millisecond * 100)
		waitTime--
	}

	assert.Equal(t, 1, len(ps1.host.Network().Peers()))
	assert.Equal(t, 1, len(ps2.host.Network().Peers()))

	pi2, err := ps2.AddrListen(ctx)
	assert.Nil(t, err)
	assert.Equal(t, pi2, peer.AddrInfo{
		ID:    ps2.host.ID(),
		Addrs: ps2.host.Addrs(),
	})

	pi1, err := ps2.FindPeer(ctx, ps1.host.ID())
	assert.Nil(t, err)
	assert.Equal(t, ps1.host.ID(), pi1.ID)

	err = ps2.Connect(ctx, pi1)
	assert.Nil(t, err)
}
