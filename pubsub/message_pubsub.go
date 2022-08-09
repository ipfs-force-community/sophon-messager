package pubsub

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/filecoin-project/venus-messager/log"
	"github.com/filecoin-project/venus/fixtures/networks"
	"github.com/filecoin-project/venus/pkg/net"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	swarm "github.com/libp2p/go-libp2p/p2p/net/swarm"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

type IMessagePubSub interface {
	IPublisher
	INet
}

type INet interface {
	Connect(ctx context.Context, p peer.AddrInfo) error
	FindPeer(ctx context.Context, peerID peer.ID) (peer.AddrInfo, error)
	Peers(ctx context.Context) ([]peer.AddrInfo, error)
	AddrListen(ctx context.Context) (peer.AddrInfo, error)
}

type IPublisher interface {
	Publish(ctx context.Context, msg *types.SignedMessage) error
}

var ErrPubsubDisabled = errors.New("pubsub is disabled")

type MessagePubSub struct {
	topic         *pubsub.Topic
	host          types.RawHost
	pubsub        *pubsub.PubSub
	log           *log.Logger
	dht           *dht.IpfsDHT
	bootstrappers []peer.AddrInfo
	period        time.Duration
	expanding     chan struct{}
}

func NewMessagePubSub(logger *log.Logger, listenAddress string, networkName types.NetworkName, bootstrap []string) (*MessagePubSub, error) {
	ctx := context.Background()

	// get default bootstrap from net params
	netconfig, err := networks.GetNetworkConfig(string(networkName))
	if err != nil {
		logger.Infof("failed to get default network config: %s", err)
	}
	if netconfig != nil {
		bootstrap = append(bootstrap, netconfig.Bootstrap.Addresses...)
	}
	rawHost, err := buildHost(ctx, listenAddress)
	if err != nil {
		return nil, err
	}

	bootstrapPeersres := make([]peer.AddrInfo, len(bootstrap))
	for i, addr := range bootstrap {
		peerInfo, err := peer.AddrInfoFromString(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse bootstrap addresses: %w", err)
		}
		bootstrapPeersres[i] = *peerInfo
	}
	router, err := makeDHT(ctx, rawHost, string(networkName), bootstrapPeersres)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %s", err)
	}

	peerHost := routedhost.Wrap(rawHost, router)

	pubsub.GossipSubHeartbeatInterval = 100 * time.Millisecond
	options := []pubsub.Option{
		// Gossipsubv1.1 configuration
		pubsub.WithFloodPublish(true),
		//  buffer, 32 -> 10K
		pubsub.WithValidateQueueSize(10 << 10),
		//  worker, 1x cpu -> 2x cpu
		pubsub.WithValidateWorkers(runtime.NumCPU() * 2),
		//  goroutine, 8K -> 16K
		pubsub.WithValidateThrottle(16 << 10),
		pubsub.WithMessageSigning(true),
	}

	gsub, err := pubsub.NewGossipSub(ctx, peerHost, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	topicName := fmt.Sprintf("/fil/msgs/%s", networkName)
	topic, err := gsub.Join(topicName)
	if err != nil {
		return nil, fmt.Errorf("failed to join topic %s: %w", topicName, err)
	}

	pubsub := MessagePubSub{
		topic:         topic,
		host:          peerHost,
		pubsub:        gsub,
		period:        5 * time.Second,
		bootstrappers: bootstrapPeersres,
		log:           logger,
		dht:           router,
		expanding:     make(chan struct{}, 1),
	}

	go pubsub.run(ctx)
	return &pubsub, nil
}

func (m *MessagePubSub) Publish(ctx context.Context, msg *types.SignedMessage) error {
	buf := new(bytes.Buffer)
	err := msg.MarshalCBOR(buf)
	if err != nil {
		return fmt.Errorf("marshal message failed %w", err)
	}

	err = m.topic.Publish(ctx, buf.Bytes())
	if err != nil {
		return fmt.Errorf("publish message failed %w", err)
	}

	m.log.Debugf("publish message %s", msg.Cid())
	return nil
}

func (m *MessagePubSub) run(ctx context.Context) {
	err := m.connectBootstrap(ctx)
	if err != nil {
		m.log.Errorf("connect bootstrap failed %s", err)
	}

	ticker := time.NewTicker(m.period)
	for range ticker.C {
		m.expandPeers()
	}
}

func (m *MessagePubSub) Connect(ctx context.Context, p peer.AddrInfo) error {
	if swarm, ok := m.host.Network().(*swarm.Swarm); ok {
		swarm.Backoff().Clear(p.ID)
	}
	return m.host.Connect(ctx, p)
}

func (m *MessagePubSub) Peers(ctx context.Context) ([]peer.AddrInfo, error) {
	if m.host == nil {
		return nil, errors.New("messager must be online")
	}

	conns := m.host.Network().Conns()
	peers := make([]peer.AddrInfo, 0, len(conns))
	for _, conn := range conns {
		peers = append(peers, peer.AddrInfo{
			ID:    conn.RemotePeer(),
			Addrs: []ma.Multiaddr{conn.RemoteMultiaddr()},
		})
	}

	return peers, nil
}

// FindPeer searches the libp2p router for a given peer id
func (m *MessagePubSub) FindPeer(ctx context.Context, peerID peer.ID) (peer.AddrInfo, error) {
	return m.dht.FindPeer(ctx, peerID)
}

func (m *MessagePubSub) AddrListen(ctx context.Context) (peer.AddrInfo, error) {
	if m.host == nil {
		return peer.AddrInfo{}, errors.New("messager must be online")
	}

	return peer.AddrInfo{
		ID:    m.host.ID(),
		Addrs: m.host.Addrs(),
	}, nil
}

func (m *MessagePubSub) connectBootstrap(ctx context.Context) error {
	for _, bsp := range m.bootstrappers {
		if err := m.host.Connect(ctx, bsp); err != nil {
			m.log.Warnf("failed to connect to bootstrap peer: %s %s", bsp.String(), err)
		}
	}
	return nil
}

func (m *MessagePubSub) expandPeers() {
	select {
	case m.expanding <- struct{}{}:
	default:
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second*30)
		defer cancel()

		m.doExpand(ctx)

		<-m.expanding
	}()
}

func (m *MessagePubSub) doExpand(ctx context.Context) {
	pcount := len(m.host.Network().Peers())
	if pcount == 0 {
		if len(m.bootstrappers) == 0 {
			m.log.Info("no peers connected, and no bootstrappers configured")
			return
		}

		m.log.Info("connecting to bootstrap peers")
		err := m.connectBootstrap(ctx)
		if err != nil {
			m.log.Info("failed to connect to bootstrap peers")
		}
		return
	}

	// if we already have some peers and need more, the dht is really good at connecting to most peers. Use that for now until something better comes along.
	if err := m.dht.Bootstrap(ctx); err != nil {
		m.log.Warnf("dht bootstrapping failed: %s", err)
	}
}

func makeDHT(ctx context.Context, h types.RawHost, networkName string, bootNodes []peer.AddrInfo) (*dht.IpfsDHT, error) {
	mode := dht.ModeAuto
	opts := []dht.Option{dht.Mode(mode),
		dht.ProtocolPrefix(net.FilecoinDHT(networkName)),
		dht.QueryFilter(dht.PublicQueryFilter),
		dht.RoutingTableFilter(dht.PublicRoutingTableFilter),
		dht.DisableProviders(),
		dht.BootstrapPeers(bootNodes...),
		dht.DisableValues()}
	r, err := dht.New(
		ctx, h, opts...,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup routing")
	}

	return r, nil
}

func buildHost(ctx context.Context, address string) (types.RawHost, error) {
	opts := []libp2p.Option{
		libp2p.UserAgent("venus-messager"),
		libp2p.ListenAddrStrings(address),
		// libp2p.Identity(secret),
		libp2p.Ping(true),
		libp2p.DisableRelay(),
	}
	return libp2p.New(opts...)
}

type MessagerPubSubStub struct {
}

func (m *MessagerPubSubStub) Publish(ctx context.Context, msg *types.SignedMessage) error {
	return ErrPubsubDisabled
}

func (m *MessagerPubSubStub) Peers(ctx context.Context) ([]peer.AddrInfo, error) {
	return nil, ErrPubsubDisabled
}

func (m *MessagerPubSubStub) Connect(ctx context.Context, p peer.AddrInfo) error {
	return ErrPubsubDisabled
}

func (m *MessagerPubSubStub) FindPeer(ctx context.Context, peerID peer.ID) (peer.AddrInfo, error) {
	return peer.AddrInfo{}, ErrPubsubDisabled
}

func (m *MessagerPubSubStub) AddrListen(ctx context.Context) (peer.AddrInfo, error) {
	return peer.AddrInfo{}, ErrPubsubDisabled
}

var _ IMessagePubSub = &MessagePubSub{}
var _ IMessagePubSub = &MessagerPubSubStub{}
