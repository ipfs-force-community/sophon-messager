package jwt

import (
	"context"

	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/venus-auth/cmd/jwtclient"
	"github.com/filecoin-project/venus-auth/core"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-messager/config"
)

type remoteJwtClient struct {
	cli *jwtclient.JWTClient
}

func newRemoteJwtClient(jwtCfg *config.JWTConfig) jwtclient.IJwtAuthClient {
	var remote *remoteJwtClient
	if len(jwtCfg.AuthURL) > 0 {
		remote = &remoteJwtClient{}
		remote.cli = jwtclient.NewJWTClient(jwtCfg.AuthURL)
	}

	return remote
}

func (c *remoteJwtClient) Verify(ctx context.Context, token string) ([]auth.Permission, error) {
	if c != nil && c.cli != nil {
		res, err := c.cli.Verify(ctx, token)
		if err != nil {
			return nil, err
		}

		return core.AdaptOldStrategy(res.Perm), nil
	}

	return nil, xerrors.New("remote client is nil")
}
