package main

import (
	"encoding/hex"
	"fmt"
	"net"
	_ "net/http/pprof"
	"os"
	"path/filepath"

	"github.com/filecoin-project/venus-messager/metrics"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-messager/api"
	"github.com/filecoin-project/venus-messager/api/jwt"
	ccli "github.com/filecoin-project/venus-messager/cli"
	"github.com/filecoin-project/venus-messager/config"
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
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "./messager.toml",
				Usage:   "specify config file",
			},
		},
		Commands: []*cli.Command{ccli.MsgCmds,
			ccli.AddrCmds,
			ccli.SharedParamsCmds,
			ccli.NodeCmds,
			ccli.LogCmds,
			ccli.SendCmd,
			runCmd,
		},
	}
	app.Version = version.Version + "--" + version.GitCommit
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
			Name:  "sqlite-file",
			Usage: "the path and file name of SQLite, eg. ~/sqlite/message.db",
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
	path := ctx.String("config")
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	exist, err := config.ConfigExist(path)
	if err != nil {
		return err
	}

	var cfg *config.Config
	if !exist {
		cfg = config.DefaultConfig()
		err = updateFlag(cfg, ctx)
		if err != nil {
			return err
		}
		if err := genSecret(&cfg.JWT); err != nil {
			return xerrors.Errorf("failed to generate secret %v", err)
		}
		err = config.WriteConfig(path, cfg)
		if err != nil {
			return err
		}
	} else {
		cfg, err = config.ReadConfig(path)
		if err != nil {
			return err
		}
		if len(cfg.JWT.Local.Secret) == 0 {
			if err := genSecret(&cfg.JWT); err != nil {
				return xerrors.Errorf("failed to generate secret %v", err)
			}
			err = config.WriteConfig(path, cfg)
			if err != nil {
				return err
			}
		}
		err = updateFlag(cfg, ctx)
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
	client, closer, err := service.NewNodeClient(ctx.Context, &cfg.Node)
	if err != nil {
		return xerrors.Errorf("connect to node failed %v", err)
	}
	defer closer()

	mAddr, err := ma.NewMultiaddr(cfg.API.Address)
	if err != nil {
		return err
	}

	var walletClient *gateway.IWalletCli
	walletCli, walletCliCloser, err := gateway.NewWalletClient(&cfg.Gateway, log)
	walletClient = &gateway.IWalletCli{IWalletClient: walletCli}
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

	shutdownChan := make(chan struct{})
	provider := fx.Options(
		fx.Logger(fxLogger{log}),
		// prover
		fx.Supply(cfg, &cfg.DB, &cfg.API, &cfg.JWT, &cfg.Node, &cfg.Log, &cfg.MessageService, &cfg.MessageState, &cfg.Gateway, &cfg.RateLimit, &cfg.Trace),
		fx.Supply(log),
		fx.Supply(client),
		fx.Supply(walletClient),
		fx.Provide(func() v1.FullNode {
			return client
		}),
		fx.Supply((ShutdownChan)(shutdownChan)),

		fx.Provide(service.NewMessageState),
		// db
		fx.Provide(models.SetDataBase),
		// service
		service.MessagerService(),
		// api
		fx.Provide(api.NewMessageImp),
		// jwt
		fx.Provide(jwt.NewJwtClient),

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
	)

	apiOption := fx.Options(
		fx.Provide(api.BindRateLimit),
		fx.Invoke(api.RunAPI),
	)

	app := fx.New(provider, invoker, apiOption)
	if err := app.Start(ctx.Context); err != nil {
		// comment fx.NopLogger few lines above for easier debugging
		return xerrors.Errorf("starting node: %w", err)
	}

	go func() {
		<-shutdownChan
		log.Warn("received shutdown")

		log.Warn("Shutting down...")
		if err := app.Stop(ctx.Context); err != nil {
			log.Errorf("graceful shutting down failed: %s", err)
		}
		log.Warn("Graceful shutdown successful")
	}()

	<-app.Done()
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
			if ctx.IsSet("sqlite-file") {
				cfg.DB.Sqlite.File = ctx.String("sqlite-file")
			}
		case "mysql":
			if ctx.IsSet("mysql-dsn") {
				cfg.DB.MySql.ConnectionString = ctx.String("mysql-dsn")
			}
		default:
			return xerrors.Errorf("unexpected db type %s", cfg.DB.Type)
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

func genSecret(cfg *config.JWTConfig) error {
	if len(cfg.Local.Secret) == 0 {
		sBytes, tBytes, err := jwt.GenSecretAndToken()
		if err != nil {
			return err
		}
		cfg.Local.Secret = hex.EncodeToString(sBytes)
		cfg.Local.Token = string(tBytes)
	}

	return nil
}
