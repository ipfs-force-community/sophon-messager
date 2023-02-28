package pubsub

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus/fixtures/networks"
	"github.com/filecoin-project/venus/pkg/net"
	"github.com/filecoin-project/venus/venus-shared/types"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	swarm "github.com/libp2p/go-libp2p/p2p/net/swarm"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

var ErrPubsubDisabled = errors.New("pubsub is disabled")
var log = logging.Logger("msg-pubsub")

type INet interface {
	Connect(ctx context.Context, p peer.AddrInfo) error
	FindPeer(ctx context.Context, peerID peer.ID) (peer.AddrInfo, error)
	Peers(ctx context.Context) ([]peer.AddrInfo, error)
	AddrListen(ctx context.Context) (peer.AddrInfo, error)
}

type IPubsuber interface {
	GetTopic(topic string) (*pubsub.Topic, error)
}

var _ INet = &PubSub{}
var _ IPubsuber = &PubSub{}

type PubSub struct {
	host             types.RawHost
	pubsub           *pubsub.PubSub
	dht              *dht.IpfsDHT
	bootstrappers    []peer.AddrInfo
	period           time.Duration
	timeout          time.Duration
	minPeerThreshold int
	expanding        chan struct{}
}

func NewPubsub(ctx context.Context,
	listenAddress string,
	networkName types.NetworkName,
	bootstrap []string,
	period time.Duration,
	threshold int,
) (*PubSub, error) {
	finalTimeout, finalPeriod, finalThreshold := time.Second*30, time.Second*30, 1

	netconfig, err := networks.GetNetworkConfigFromName(string(networkName))
	if err != nil {
		log.Errorf("failed to get default network config: %s", err)
	}
	if netconfig != nil {
		bootstrap = append(bootstrap, netconfig.Bootstrap.Addresses...)
		finalThreshold = netconfig.Bootstrap.MinPeerThreshold
		_ = toml.Unmarshal([]byte(netconfig.Bootstrap.Period), &finalPeriod)
	}
	if period != 0 {
		finalPeriod = period
	}
	if threshold != 0 {
		finalThreshold = threshold
	}
	if finalTimeout > finalPeriod {
		finalTimeout = finalPeriod
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

	pubsub := PubSub{
		host:             peerHost,
		pubsub:           gsub,
		bootstrappers:    bootstrapPeersres,
		dht:              router,
		expanding:        make(chan struct{}, 1),
		minPeerThreshold: finalThreshold,
		period:           finalPeriod,
		timeout:          finalTimeout,
	}

	go pubsub.run(ctx)
	return &pubsub, nil
}

func (m *PubSub) GetTopic(topic string) (*pubsub.Topic, error) {
	return m.pubsub.Join(topic)
}

func (m *PubSub) run(ctx context.Context) {
	err := m.connectBootstrap(ctx)
	if err != nil {
		log.Errorf("connect bootstrap failed %s", err)
	}

	ticker := time.NewTicker(m.period)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			pcount := len(m.host.Network().Peers())
			if pcount <= m.minPeerThreshold {
				log.Debug("peer count %d is less than threshold %d, expanding", pcount, m.minPeerThreshold)
				m.expandPeers()
			}

		case <-ctx.Done():
			log.Warnf("stop expand peers: %v", ctx.Err())
			return
		}
	}
}

func (m *PubSub) Connect(ctx context.Context, p peer.AddrInfo) error {
	if swarm, ok := m.host.Network().(*swarm.Swarm); ok {
		swarm.Backoff().Clear(p.ID)
	}
	return m.host.Connect(ctx, p)
}

func (m *PubSub) Peers(ctx context.Context) ([]peer.AddrInfo, error) {
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
func (m *PubSub) FindPeer(ctx context.Context, peerID peer.ID) (peer.AddrInfo, error) {
	return m.dht.FindPeer(ctx, peerID)
}

func (m *PubSub) AddrListen(ctx context.Context) (peer.AddrInfo, error) {
	return peer.AddrInfo{
		ID:    m.host.ID(),
		Addrs: m.host.Addrs(),
	}, nil
}

func (m *PubSub) connectBootstrap(ctx context.Context) error {
	for _, bsp := range m.bootstrappers {
		if err := m.host.Connect(ctx, bsp); err != nil {
			log.Warnf("failed to connect to bootstrap peer: %s %s", bsp, err)
		}
	}
	return nil
}

func (m *PubSub) expandPeers() {
	select {
	case m.expanding <- struct{}{}:
	default:
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.TODO(), m.timeout)
		defer cancel()

		m.doExpand(ctx)

		<-m.expanding
	}()
}

func (m *PubSub) doExpand(ctx context.Context) {
	pcount := len(m.host.Network().Peers())
	if pcount == 0 {
		if len(m.bootstrappers) == 0 {
			log.Info("no peers connected, and no bootstrappers configured")
			return
		}

		log.Info("connecting to bootstrap peers")
		err := m.connectBootstrap(ctx)
		if err != nil {
			log.Info("failed to connect to bootstrap peers")
		}
		return
	}

	// if we already have some peers and need more, the dht is really good at connecting to most peers. Use that for now until something better comes along.
	if err := m.dht.Bootstrap(ctx); err != nil {
		log.Warnf("dht bootstrapping failed: %s", err)
	}
}

func makeDHT(ctx context.Context, h types.RawHost, networkName string, bootNodes []peer.AddrInfo) (*dht.IpfsDHT, error) {
	mode := dht.ModeAuto
	opts := []dht.Option{
		dht.Mode(mode),
		dht.ProtocolPrefix(net.FilecoinDHT(networkName)),
		dht.QueryFilter(dht.PublicQueryFilter),
		dht.RoutingTableFilter(dht.PublicRoutingTableFilter),
		dht.DisableProviders(),
		dht.BootstrapPeers(bootNodes...),
		dht.DisableValues(),
	}
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

func ProvidePubsub(ctx context.Context, networkName types.NetworkName, net *config.Libp2pNetConfig) (*PubSub, error) {
	return NewPubsub(ctx, net.ListenAddress, networkName, net.BootstrapAddresses, net.ExpandPeriod, net.MinPeerThreshold)
}

func NewINet(p *PubSub) INet {
	return p
}

func NewIPubsuber(p *PubSub) IPubsuber {
	return p
}

func Options() fx.Option {
	return fx.Options(
		fx.Provide(ProvidePubsub),
		fx.Provide(NewINet),
		fx.Provide(NewIPubsuber),
	)
}
