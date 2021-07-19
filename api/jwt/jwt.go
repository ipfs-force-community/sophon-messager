package jwt

import (
	"github.com/filecoin-project/venus-auth/cmd/jwtclient"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-messager/config"
)

type JwtClient struct {
	Local  jwtclient.IJwtAuthClient
	Remote *RemoteAuthClient
}

func NewJwtClient(jwtCfg *config.JWTConfig) (*JwtClient, error) {
	var err error
	jc := &JwtClient{
		Remote: newRemoteJwtClient(jwtCfg),
	}
	if jc.Local, err = newLocalJWTClient(jwtCfg); err != nil {
		return nil, xerrors.Errorf("new local jwt client failed %v", err)
	}

	return jc, nil
}
