package jwt

import (
	"github.com/ipfs-force-community/venus-auth/cmd/jwtclient"
	"github.com/ipfs-force-community/venus-messager/config"
)

func NewJwtClient(jwtCfg *config.JWTConfig) IJwtClient {
	if len(jwtCfg.Url) > 0 {
		return jwtclient.NewJWTClient(jwtCfg.Url)
	}
	return LocalClient{}
}
