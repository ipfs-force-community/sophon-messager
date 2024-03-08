package api

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/etherlabsio/healthcheck/v2"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/filecoin-project/venus/venus-shared/api/permission"
	"github.com/ipfs-force-community/metrics/ratelimit"
	"github.com/ipfs-force-community/sophon-auth/core"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"
	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"

	"github.com/ipfs-force-community/sophon-messager/config"
)

var log = logging.Logger("api")

func BindRateLimit(msgImp *MessageImp, remoteAuthCli jwtclient.IAuthClient, rateLimitCfg *config.RateLimitConfig) (messager.IMessager, error) {
	var msgAPI messager.IMessagerStruct
	permission.PermissionProxy(msgImp, &msgAPI)

	if len(rateLimitCfg.Redis) != 0 && remoteAuthCli != nil {
		limiter, err := ratelimit.NewRateLimitHandler(
			rateLimitCfg.Redis,
			nil,
			&core.ValueFromCtx{},
			jwtclient.WarpLimitFinder(remoteAuthCli),
			logging.Logger("rate-limit"),
		)
		if err != nil {
			return nil, err
		}
		var rateLimitAPI messager.IMessagerStruct
		limiter.WraperLimiter(msgAPI.Internal, &rateLimitAPI.Internal)
		msgAPI = rateLimitAPI
	}
	return &msgAPI, nil
}

// RunAPI bind rpc call and start rpc
// todo
func RunAPI(lc fx.Lifecycle, localAuthCli *jwtclient.LocalAuthClient, remoteAuthCli jwtclient.IAuthClient, lst net.Listener, msgImp messager.IMessager) error {
	srv := jsonrpc.NewServer()
	srv.Register("Message", msgImp)
	authMux := jwtclient.NewAuthMux(localAuthCli, jwtclient.WarpIJwtAuthClient(remoteAuthCli), srv)

	mux := http.NewServeMux()
	mux.Handle("/rpc/v0", authMux)
	mux.Handle("/debug/pprof/", http.DefaultServeMux)
	mux.Handle("/healthcheck", healthcheck.Handler())

	apiserv := &http.Server{
		Handler: mux,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Info("start rpcserver ", lst.Addr())
				core.ApiState.Set(ctx, 1)
				if err := apiserv.Serve(lst); err != nil && !errors.Is(err, http.ErrServerClosed) {
					core.ApiState.Set(ctx, 0)
					log.Errorf("start rpcserver failed: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			defer core.ApiState.Set(ctx, 0)
			return apiserv.Shutdown(ctx)
		},
	})
	return nil
}
