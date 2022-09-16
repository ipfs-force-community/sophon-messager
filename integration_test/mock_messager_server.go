package integration

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/filecoin-project/venus-auth/jwtclient"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/venus/venus-shared/api/messager"

	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	gatewayapi "github.com/filecoin-project/venus/venus-shared/api/gateway/v1"
	"github.com/mitchellh/go-homedir"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"go.uber.org/fx"

	"github.com/filecoin-project/venus-messager/api"
	ccli "github.com/filecoin-project/venus-messager/cli"
	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/filestore"
	"github.com/filecoin-project/venus-messager/gateway"
	"github.com/filecoin-project/venus-messager/log"
	"github.com/filecoin-project/venus-messager/metrics"
	"github.com/filecoin-project/venus-messager/models"
	"github.com/filecoin-project/venus-messager/service"
	"github.com/filecoin-project/venus-messager/testhelper"
)

type messagerServer struct {
	log       *log.Logger
	walletCli *gateway.MockWalletProxy
	fullNode  *testhelper.MockFullNode

	token string
	port  string

	app         *fx.App
	appStartErr chan error
}

func mockMessagerServer(ctx context.Context, repoPath string, cfg *config.Config) (*messagerServer, error) {
	repoPath, err := homedir.Expand(repoPath)
	if err != nil {
		return nil, err
	}

	remoteAuthCli := &jwtclient.AuthClient{}

	localAuthCli, token, err := jwtclient.NewLocalAuthClient()
	if err != nil {
		return nil, fmt.Errorf("failed to generate local auth client %v", err)
	}

	fsRepo, err := filestore.InitFSRepo(repoPath, cfg)
	if err != nil {
		return nil, err
	}

	log, err := log.SetLogger(&cfg.Log)
	if err != nil {
		return nil, err
	}

	log.Infof("node info url: %s, token: %s\n", cfg.Node.Url, cfg.Node.Token)
	log.Infof("auth info url: %s\n", cfg.JWT.AuthURL)
	log.Infof("gateway info url: %s, token: %s\n", cfg.Gateway.Url, cfg.Node.Token)
	log.Infof("rate limit info: redis: %s \n", cfg.RateLimit.Redis)

	fullNode, err := testhelper.NewMockFullNode(ctx, cfg.MessageService.WaitingChainHeadStableDuration*2)
	if err != nil {
		return nil, err
	}
	if err := ccli.LoadBuiltinActors(ctx, fullNode); err != nil {
		return nil, err
	}

	networkName, err := fullNode.StateNetworkName(ctx)
	if err != nil {
		return nil, fmt.Errorf("get network name failed %v", err)
	}

	mAddr, err := ma.NewMultiaddr(cfg.API.Address)
	if err != nil {
		return nil, err
	}

	walletCli := gateway.NewMockWalletProxy()

	// Listen on the configured address in order to bind the port number in case it has
	// been configured as zero (i.e. OS-provided)
	apiListener, err := manet.Listen(mAddr)
	if err != nil {
		return nil, err
	}
	lst := manet.NetListener(apiListener)

	provider := fx.Options(
		// prover
		fx.Supply(cfg, &cfg.DB, &cfg.API, &cfg.JWT, &cfg.Node, &cfg.Log, &cfg.MessageService, cfg.Libp2pNetConfig,
			&cfg.MessageState, &cfg.Gateway, &cfg.RateLimit, cfg.Trace, cfg.Metrics),
		fx.Supply(log),
		fx.Supply(fullNode),
		fx.Supply(networkName),
		fx.Supply(remoteAuthCli),
		fx.Supply(localAuthCli),
		fx.Provide(func() gatewayapi.IWalletClient {
			return walletCli
		}),
		fx.Provide(func() v1.FullNode {
			return fullNode
		}),
		fx.Provide(func() filestore.FSRepo {
			return fsRepo
		}),

		// db
		fx.Provide(models.SetDataBase),
		// service
		service.MessagerService(),
		// api
		fx.Provide(api.NewMessageImp),

		fx.Provide(func() context.Context {
			return ctx
		}),

		// middleware

		fx.Provide(func() net.Listener {
			return lst
		}),
	)

	invoker := fx.Options(
		// invoke
		fx.Invoke(models.AutoMigrate),
		fx.Invoke(service.StartNodeEvents),
		fx.Invoke(metrics.SetupJaeger),
		fx.Invoke(metrics.SetupMetrics),
	)

	apiOption := fx.Options(
		fx.Provide(api.BindRateLimit),
		fx.Invoke(api.RunAPI),
	)

	app := fx.New(provider, invoker, apiOption)

	return &messagerServer{
		log:         log,
		walletCli:   walletCli,
		fullNode:    fullNode,
		token:       string(token),
		port:        strings.Split(lst.Addr().String(), ":")[1],
		app:         app,
		appStartErr: make(chan error),
	}, nil
}

func (ms *messagerServer) start(ctx context.Context) {
	ms.appStartErr <- ms.app.Start(ctx)
}

func (ms *messagerServer) stop(ctx context.Context) error {
	return ms.app.Stop(ctx)
}

func newMessagerClient(ctx context.Context, port, token string) (messager.IMessager, jsonrpc.ClientCloser, error) {
	return messager.DialIMessagerRPC(ctx, fmt.Sprintf("/ip4/127.0.0.1/tcp/%s", port), token, nil)
}
