package main

import (
	"errors"
	"fmt"
	"net"
	_ "net/http/pprof"
	"os"

	"github.com/filecoin-project/venus-messager/metrics"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	gatewayapi "github.com/filecoin-project/venus/venus-shared/api/gateway/v1"
	"github.com/mitchellh/go-homedir"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	"github.com/urfave/cli/v2"
	"go.uber.org/fx"

	"github.com/filecoin-project/venus-messager/api"

	"github.com/filecoin-project/venus-auth/jwtclient"
	ccli "github.com/filecoin-project/venus-messager/cli"
	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/filestore"
	"github.com/filecoin-project/venus-messager/gateway"
	"github.com/filecoin-project/venus-messager/log"
	"github.com/filecoin-project/venus-messager/models"
	"github.com/filecoin-project/venus-messager/service"
	"github.com/filecoin-project/venus-messager/version"
)

func main() {
	app := &cli.App{
		Name:  "venus message",
		Usage: "used for manage message",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "repo",
				Value: "~/.venus-messager",
			},
		},
		Commands: []*cli.Command{ccli.MsgCmds,
			ccli.AddrCmds,
			ccli.SharedParamsCmds,
			ccli.NodeCmds,
			ccli.LogCmds,
			ccli.SendCmd,
			ccli.SwarmCmds,
			runCmd,
		},
	}

	app.Version = version.Version
	app.Setup()
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		return
	}

}

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "run messager",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "auth-url",
			Usage: "url for auth server",
		},

		// node
		&cli.StringFlag{
			Name:  "node-url",
			Usage: "url for connection lotus/venus",
		},
		&cli.StringFlag{
			Name:  "node-token",
			Usage: "token auth for lotus/venus",
		},

		// database
		&cli.StringFlag{
			Name:  "db-type",
			Usage: "which db to use. sqlite/mysql",
		},
		&cli.StringFlag{
			Name:  "mysql-dsn",
			Usage: "mysql connection string",
		},
		&cli.StringSliceFlag{
			Name:  "gateway-url",
			Usage: "gateway url",
		},
		&cli.StringFlag{
			Name:  "gateway-token",
			Usage: "gateway token",
		},
		&cli.StringFlag{
			Name:  "auth-token",
			Usage: "auth token",
		},
		&cli.StringFlag{
			Name: "rate-limit-redis",
		},
	},
	Action: runAction,
}

func runAction(ctx *cli.Context) error {
	var fsRepo filestore.FSRepo
	cfg := config.DefaultConfig()

	repoPath, err := homedir.Expand(ctx.String("repo"))
	if err != nil {
		return err
	}
	hasFSRepo, err := hasFSRepo(repoPath)
	if err != nil {
		return err
	}
	if hasFSRepo {
		fsRepo, err = filestore.NewFSRepo(repoPath)
		if err != nil {
			return err
		}
		cfg = fsRepo.Config()
	}

	if err = updateFlag(cfg, ctx); err != nil {
		return err
	}

	if !hasFSRepo {
		fsRepo, err = filestore.InitFSRepo(repoPath, cfg)
		if err != nil {
			return err
		}
	}

	log, err := log.SetLogger(&cfg.Log)
	if err != nil {
		return err
	}

	log.Infof("node info url: %s, token: %s\n", cfg.Node.Url, cfg.Node.Token)
	log.Infof("auth info url: %s\n", cfg.JWT.AuthURL)
	log.Infof("gateway info url: %s, token: %s\n", cfg.Gateway.Url, cfg.Node.Token)
	log.Infof("rate limit info: redis: %s \n", cfg.RateLimit.Redis)

	remoteAuthCli, err := jwtclient.NewAuthClient(cfg.JWT.AuthURL)
	if err != nil {
		return err
	}

	localAuthCli, token, err := jwtclient.NewLocalAuthClient()
	if err != nil {
		return fmt.Errorf("failed to generate local auth client %v", err)
	}

	fsRepo.SaveToken(token)

	client, closer, err := v1.DialFullNodeRPC(ctx.Context, cfg.Node.Url, cfg.Node.Token, nil)
	if err != nil {
		return fmt.Errorf("connect to node failed %v", err)
	}
	defer closer()

	networkName, err := client.StateNetworkName(ctx.Context)
	if err != nil {
		return fmt.Errorf("get network name failed %v", err)
	}

	if err := ccli.LoadBuiltinActors(ctx.Context, client); err != nil {
		return err
	}

	mAddr, err := ma.NewMultiaddr(cfg.API.Address)
	if err != nil {
		return err
	}

	walletCli, walletCliCloser, err := gateway.NewWalletClient(&cfg.Gateway, log)
	if err != nil {
		return err
	}
	defer walletCliCloser()

	// Listen on the configured address in order to bind the port number in case it has
	// been configured as zero (i.e. OS-provided)
	apiListener, err := manet.Listen(mAddr)
	if err != nil {
		return err
	}
	lst := manet.NetListener(apiListener)

	provider := fx.Options(
		fx.Logger(fxLogger{log}),
		// prover
		fx.Supply(cfg, &cfg.DB, &cfg.API, &cfg.JWT, &cfg.Node, &cfg.Log, &cfg.MessageService, cfg.Libp2pNetConfig,
			&cfg.MessageState, &cfg.Gateway, &cfg.RateLimit, cfg.Trace, cfg.Metrics),
		fx.Supply(log),
		fx.Supply(client),
		fx.Supply(networkName),
		fx.Supply(remoteAuthCli),
		fx.Supply(localAuthCli),
		fx.Provide(func() gatewayapi.IWalletClient {
			return walletCli
		}),
		fx.Provide(func() v1.FullNode {
			return client
		}),
		fx.Provide(func() filestore.FSRepo {
			return fsRepo
		}),

		fx.Provide(service.NewMessageState),
		// db
		fx.Provide(models.SetDataBase),
		// service
		service.MessagerService(),
		// api
		fx.Provide(api.NewMessageImp),

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
	if err := app.Start(ctx.Context); err != nil {
		// comment fx.NopLogger few lines above for easier debugging
		return fmt.Errorf("starting app: %w", err)
	}

	shutdownChan := make(chan struct{})
	// wait for exit to complete
	finishCh := make(chan struct{})
	go func() {
		<-shutdownChan

		log.Warn("received shutdown")

		log.Warn("Shutting down...")
		if err := app.Stop(ctx.Context); err != nil {
			log.Errorf("graceful shutting down failed: %s", err)
		}
		log.Info("Graceful shutdown successful")

		close(finishCh)
	}()

	<-app.Done()

	shutdownChan <- struct{}{}

	<-finishCh

	return nil
}

func updateFlag(cfg *config.Config, ctx *cli.Context) error {
	if ctx.IsSet("auth-url") {
		cfg.JWT.AuthURL = ctx.String("auth-url")
	}

	if ctx.IsSet("node-url") {
		cfg.Node.Url = ctx.String("node-url")
	}

	if ctx.IsSet("gateway-url") {
		cfg.Gateway.Url = ctx.StringSlice("gateway-url")
	}

	if ctx.IsSet("auth-token") {
		cfg.Node.Token = ctx.String("auth-token")
		cfg.Gateway.Token = ctx.String("auth-token")
	}

	if ctx.IsSet("node-token") {
		cfg.Node.Token = ctx.String("node-token")
	}

	if ctx.IsSet("gateway-token") {
		cfg.Gateway.Token = ctx.String("gateway-token")
	}

	if ctx.IsSet("db-type") {
		cfg.DB.Type = ctx.String("db-type")
		switch cfg.DB.Type {
		case "sqlite":
		case "mysql":
			if ctx.IsSet("mysql-dsn") {
				cfg.DB.MySql.ConnectionString = ctx.String("mysql-dsn")
			}
		default:
			return fmt.Errorf("unexpected db type %s", cfg.DB.Type)
		}
	}
	if ctx.IsSet("rate-limit-redis") {
		cfg.RateLimit.Redis = ctx.String("rate-limit-redis")
	}
	return nil
}

type fxLogger struct {
	log *log.Logger
}

func (l fxLogger) Printf(str string, args ...interface{}) {
	l.log.Infof(str, args...)
}

func hasFSRepo(repoPath string) (bool, error) {
	fi, err := os.Stat(repoPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	if !fi.IsDir() {
		return false, fmt.Errorf("%s is not a folder", repoPath)
	}

	return true, nil
}
