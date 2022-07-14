package jwt

import (
	"fmt"

	"github.com/filecoin-project/venus-auth/cmd/jwtclient"

	"github.com/filecoin-project/venus-messager/config"
)

type Client struct {
	Local  jwtclient.IJwtAuthClient
	Remote *RemoteAuthClient
}

func NewJwtClient(jwtCfg *config.JWTConfig) (*Client, error) {
	var err error
	jc := &Client{
		Remote: newRemoteJwtClient(jwtCfg),
	}
	if jc.Local, err = newLocalJWTClient(jwtCfg); err != nil {
		return nil, fmt.Errorf("new local jwt client failed %v", err)
	}

	return jc, nil
}
