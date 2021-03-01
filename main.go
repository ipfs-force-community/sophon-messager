package main

import (
	"fmt"
	"github.com/ipfs-force-community/venus-messager/api"
	"github.com/ipfs-force-community/venus-messager/api/controller"
	msgCli "github.com/ipfs-force-community/venus-messager/cli"
	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
	"golang.org/x/xerrors"
	"net"
	"os"
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
		Commands: []*cli.Command{msgCli.MsgCmds},
	}
	app.Setup()
	app.Action = runAction
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		return
	}

}

func runAction(ctx *cli.Context) error {
	path := ctx.String("config")

	cfg, err := config.ReadConfig(path)
	if err != nil {
		return err
	}

	log, err := SetLogger(&cfg.Log)
	if err != nil {
		return err
	}

	lst, err := net.Listen("tcp", cfg.API.Address)

	shutdownChan := make(chan struct{})
	provider := fx.Options(
		//prover
		fx.Supply((ShutdownChan)(shutdownChan), log, cfg, &cfg.DbConfig, &cfg.API, &cfg.JWT),
		fx.Provide(models.SetDataBase),
		fx.Provide(api.InitRouter),
		fx.Provide(func() net.Listener {
			return lst
		}),
	)

	invoker := fx.Options(
		//invoke
		fx.Invoke(models.AutoMigrate),
		fx.Invoke(api.UseMiddleware),
		fx.Invoke(controller.SetupController),
		fx.Invoke(api.RunAPI),
	)
	app := fx.New(provider, invoker)
	if err := app.Start(ctx.Context); err != nil {
		// comment fx.NopLogger few lines above for easier debugging
		return xerrors.Errorf("starting node: %w", err)
	}

	go func() {
		select {
		case <-shutdownChan:
			log.Warn("received shutdown")
		}

		log.Warn("Shutting down...")
		if err := app.Stop(ctx.Context); err != nil {
			log.Errorf("graceful shutting down failed: %s", err)
		}
		log.Warn("Graceful shutdown successful")
	}()

	<-app.Done()
	return nil
}
