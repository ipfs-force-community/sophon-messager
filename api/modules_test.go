package api_test

import (
	"context"
	"testing"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"
	"github.com/ipfs-force-community/sophon-messager/api"
	"github.com/ipfs-force-community/sophon-messager/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

// this test is just for checking
// weather `limiter.WraperLimiter` could warp 'msgAPI' to 'ratelimitAPI' or not,
// if ok, all member function in 'ratelimitAPI' is not nil
func TestLimitWrap(t *testing.T) {
	// todo make test injection more generic
	provider := fx.Options(
		fx.Supply(&config.RateLimitConfig{
			Redis: "test url",
		}),
		fx.Provide(func() jwtclient.IAuthClient {
			return &jwtclient.AuthClient{}
		}),
		fx.Supply(&api.MessageImp{}),
		fx.Provide(api.BindRateLimit),
	)
	app := fx.New(provider, fx.Invoke(func(_ messager.IMessager) error { return nil }))
	assert.Nil(t, app.Start(context.Background()))
}
