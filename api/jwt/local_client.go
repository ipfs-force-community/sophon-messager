package jwt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"io/ioutil"

	"github.com/filecoin-project/venus-auth/cmd/jwtclient"

	auth2 "github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/core"
	jwt3 "github.com/gbrlsnchs/jwt/v3"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-messager/config"
)

type localJwtClient struct {
	alg *jwt3.HMACSHA
}

func newLocalJWTClient(cfg *config.JWTConfig) (jwtclient.IJwtAuthClient, error) {
	lc := &localJwtClient{}

	if len(cfg.Local.Secret) == 0 {
		return nil, xerrors.Errorf("secret is empty")
	}
	b, err := hex.DecodeString(cfg.Local.Secret)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode secret %v", err)
	}
	lc.alg = jwt3.NewHS256(b)

	return lc, nil
}

func (c *localJwtClient) Verify(ctx context.Context, token string) ([]auth2.Permission, error) {
	var payload auth.JWTPayload
	_, err := jwt3.Verify([]byte(token), c.alg, &payload)
	if err != nil {
		return nil, err
	}
	return core.AdaptOldStrategy(payload.Perm), nil
}

func (c *localJwtClient) NewAuth(payload auth.JWTPayload) ([]byte, error) {
	return jwt3.Sign(payload, c.alg)
}

func GenSecretAndToken() ([]byte, []byte, error) {
	sk, err := ioutil.ReadAll(io.LimitReader(rand.Reader, 32))
	if err != nil {
		return nil, nil, err
	}

	token, err := jwt3.Sign(auth.JWTPayload{
		Name: "admin",
		Perm: "admin",
	}, jwt3.NewHS256(sk))
	if err != nil {
		return nil, nil, err
	}

	return sk, token, nil
}
