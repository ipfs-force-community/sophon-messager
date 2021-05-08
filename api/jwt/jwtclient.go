package jwt

import (
	"github.com/filecoin-project/venus-auth/cmd/jwtclient"
	"github.com/filecoin-project/venus-messager/config"
)

func NewJwtClient(jwtCfg *config.JWTConfig) IJwtClient {
	if len(jwtCfg.Url) > 0 {
		return jwtclient.NewJWTClient(jwtCfg.Url)
	}
	return LocalClient{}
}
