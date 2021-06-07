package jwt

import (
	"context"
	"net/http"
	"strings"

	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/util"
	"github.com/ipfs-force-community/venus-gateway/types"

	"github.com/filecoin-project/venus-messager/log"
)

type AuthMux struct {
	jwtClient   IJwtClient
	log         *log.Logger
	mux         *http.ServeMux
	trustHandle map[string]http.Handler
}

func NewAuthMux(jwtClient IJwtClient, logger *log.Logger, mux *http.ServeMux) *AuthMux {
	return &AuthMux{jwtClient: jwtClient, log: logger, mux: mux, trustHandle: map[string]http.Handler{}}
}

func (authMux *AuthMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handle, ok := authMux.trustHandle[r.RequestURI]; ok {
		handle.ServeHTTP(w, r)
		return
	}

	ctx := r.Context()

	token := r.Header.Get("Authorization")
	if token == "" {
		token = r.FormValue("token")
		if token != "" {
			token = "Bearer " + token
		}
	}

	if !strings.HasPrefix(token, "Bearer ") {
		authMux.log.Warn("missing Bearer prefix in header")
		w.WriteHeader(401)
		return
	}

	token = strings.TrimPrefix(token, "Bearer ")
	res, err := authMux.jwtClient.Verify(util.MacAddr(), "venus-message", r.RemoteAddr, r.Host, token)
	if err != nil {
		authMux.log.Warnf("JWT Verification failed (originating from %s): %s", r.RemoteAddr, err)
		w.WriteHeader(401)
		return
	}
	ctx = context.WithValue(ctx, types.AccountKey, res.Name)

	remoteAddr := strings.Split(r.RemoteAddr, ":")[0]
	ctx = context.WithValue(ctx, types.IPKey, remoteAddr)
	if remoteAddr == "127.0.0.1" {
		// if other nodes on the same PC, the permission check will passes directly
		ctx = core.WithPerm(ctx, core.PermAdmin)
	} else {
		ctx = core.WithPerm(ctx, res.Perm)
	}

	*r = *(r.WithContext(ctx))
	authMux.mux.ServeHTTP(w, r)
}

func (authMux *AuthMux) TruthHandle(pattern string, handle http.Handler) {
	authMux.trustHandle[pattern] = handle
}
