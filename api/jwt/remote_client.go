package jwt

import (
	"context"

	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/venus-auth/cmd/jwtclient"
	"github.com/filecoin-project/venus-auth/core"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-messager/config"
)

type RemoteAuthClient struct {
	Cli *jwtclient.JWTClient
}

func newRemoteJwtClient(jwtCfg *config.JWTConfig) *RemoteAuthClient {
	var remote *RemoteAuthClient
	if len(jwtCfg.AuthURL) > 0 {
		remote = &RemoteAuthClient{}
		remote.Cli = jwtclient.NewJWTClient(jwtCfg.AuthURL)
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

	return nil, xerrors.New("remote client is nil")
}
