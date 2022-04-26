package api_test

import (
	"context"
	"testing"

	"github.com/filecoin-project/venus-auth/cmd/jwtclient"
	"github.com/filecoin-project/venus-messager/api"
	"github.com/filecoin-project/venus-messager/api/jwt"
	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/log"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

// this test is just for checking
// weather `limiter.WraperLimiter` could warp 'msgAPI' to 'ratelimitAPI' or not,
// if ok, all member function in 'ratelimitAPI' is not nil
func TestLimitWrap(t *testing.T) {
	//todo make test injection more generic
	cfg := &config.RateLimitConfig{
		Redis: "test url",
	}
	mlog := log.New()
	//	limiter, err := ratelimit.NewRateLimitHandler(
	//		"", nil, &jwtclient.ValueFromCtx{}, jwtclient.WarpLimitFinder(cli), nil)

	provider := fx.Options(
		fx.Supply(cfg),
		fx.Provide(func() *jwt.JwtClient {
			return &jwt.JwtClient{
				Remote: &jwt.RemoteAuthClient{
					Cli: &jwtclient.AuthClient{},
				},
			}
		}),
		fx.Supply(mlog),
		fx.Supply(&api.MessageImp{}),
		fx.Provide(api.BindRateLimit),
	)
	app := fx.New(provider, fx.Invoke(func(_ messager.IMessager) error { return nil }))
	assert.Nil(t, app.Start(context.Background()))
}
