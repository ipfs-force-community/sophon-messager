package jwt

import "github.com/filecoin-project/venus-auth/auth"

type LocalClient struct {
}

func (m LocalClient) Verify(spanId, serviceName, preHost, host, token string) (*auth.VerifyResponse, error) {
	return &auth.VerifyResponse{
		Name: "admin",
		Perm: "sign",
	}, nil
}

type IJwtClient interface {
	Verify(spanId, serviceName, preHost, host, token string) (*auth.VerifyResponse, error)
}
