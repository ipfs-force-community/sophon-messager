package jwt

import (
	"context"
	"errors"

	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/venus-auth/cmd/jwtclient"
	"github.com/filecoin-project/venus-auth/core"

	"github.com/filecoin-project/venus-messager/config"
)

type RemoteAuthClient struct {
	Cli *jwtclient.AuthClient
}

func newRemoteJwtClient(jwtCfg *config.JWTConfig) *RemoteAuthClient {
	var remote *RemoteAuthClient
	if len(jwtCfg.AuthURL) > 0 {
		remote = &RemoteAuthClient{}
		remote.Cli, _ = jwtclient.NewAuthClient(jwtCfg.AuthURL)
	}

	return remote
}

func (c *RemoteAuthClient) Verify(ctx context.Context, token string) ([]auth.Permission, error) {
	if c != nil && c.Cli != nil {
		res, err := c.Cli.Verify(ctx, token)
		if err != nil {
			return nil, err
		}

		return core.AdaptOldStrategy(res.Perm), nil
	}

	return nil, errors.New("remote client is nil")
}
