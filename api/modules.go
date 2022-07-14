package api

import (
	"context"
	"net"
	"net/http"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/filecoin-project/venus/venus-shared/api/permission"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/venus-auth/cmd/jwtclient"
	"github.com/filecoin-project/venus-messager/api/jwt"
	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/log"
	"github.com/ipfs-force-community/metrics/ratelimit"
	"go.uber.org/fx"
)

func BindRateLimit(msgImp *MessageImp, jwtCli *jwt.Client, log *log.Logger, rateLimitCfg *config.RateLimitConfig) (messager.IMessager, error) {
	var msgAPI messager.IMessagerStruct
	permission.PermissionProxy(msgImp, &msgAPI)

	if len(rateLimitCfg.Redis) != 0 && jwtCli.Remote != nil && jwtCli.Remote.Cli != nil {
		limiter, err := ratelimit.NewRateLimitHandler(
			rateLimitCfg.Redis,
			nil,
			&jwtclient.ValueFromCtx{},
			jwtclient.WarpLimitFinder(jwtCli.Remote.Cli),
			log,
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
func RunAPI(lc fx.Lifecycle, jwtCli *jwt.Client, lst net.Listener, log *log.Logger, msgImp messager.IMessager) error {
	srv := jsonrpc.NewServer()
	srv.Register("Message", msgImp)
	handler := http.NewServeMux()
	handler.Handle("/rpc/v0", srv)
	authMux := jwtclient.NewAuthMux(jwtCli.Local, jwtCli.Remote, handler)
	authMux.TrustHandle("/debug/pprof/", http.DefaultServeMux)

	apiserv := &http.Server{
		Handler: authMux,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Info("Start rpcserver ", lst.Addr())
				if err := apiserv.Serve(lst); err != nil {
					log.Errorf("Start rpcserver failed: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return lst.Close()
		},
	})
	return nil
}
