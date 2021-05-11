package controller

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/util"
	"github.com/filecoin-project/venus-messager/api/jwt"
	"github.com/filecoin-project/venus-messager/types"
	"github.com/filecoin-project/venus-messager/utils"
)

type JWTFilter struct {
	jwtClient jwt.IJwtClient
	log       *logrus.Logger
	r         *gin.Engine
}

func NewJWTFilter(jwtClient jwt.IJwtClient, log *logrus.Logger, r *gin.Engine) *JWTFilter {
	return &JWTFilter{jwtClient: jwtClient, log: log, r: r}
}

func (jwtFilter *JWTFilter) PreRequest(w http.ResponseWriter, req *http.Request) (int, error) {
	localIp := utils.GetLocalIP()
	//	r.Header.get("Remote_addr")
	ip := req.Header.Get("X-Real-IP")
	if len(ip) == 0 {
		ip = strings.Split(req.RemoteAddr, ":")[0]
	}

	if len(ip) == 0 {
		return http.StatusNonAuthoritativeInfo, xerrors.New("cant get client ip")
	}

	if ip == "127.0.0.1" {
		ctx := context.WithValue(req.Context(), types.WalletInfo{}, types.WalletInfo{
			NeedCompare: false,
		})
		newReq := req.WithContext(ctx)
		*req = *newReq
		return 0, nil
	}

	token := req.Header.Get("Authorization")
	if token == "" {
		token = req.FormValue("token")
		if token != "" {
			token = "Bearer " + token
		}
	}

	if token != "" {
		if !strings.HasPrefix(token, "Bearer ") {
			return http.StatusUnauthorized, xerrors.New("missing Bearer prefix in auth header")
		}
		token = strings.TrimPrefix(token, "Bearer ")
		allow, err := jwtFilter.jwtClient.Verify(util.MacAddr(), "venus-messager", ip, localIp, token)
		if err != nil {
			return http.StatusUnauthorized, xerrors.Errorf("JWT Verification failed (originating from %s): %s", ip, err)
		}
		args, ok := req.Context().Value(types.Arguments{}).(map[string]interface{})
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return http.StatusUnauthorized, xerrors.Errorf("Not found arguments")
		}
		method := args["method"].(string)

		perms := core.AdaptOldStrategy(allow.Perm)
		if !utils.Contains(perms, authMap[method]) {
			w.WriteHeader(http.StatusUnauthorized)
			return http.StatusUnauthorized, xerrors.Errorf("Perm failed (need %s): %s", authMap[method], allow.Perm)
		}

		ctx := context.WithValue(req.Context(), types.WalletInfo{}, types.WalletInfo{
			WalletName:  allow.Name,
			NeedCompare: true,
		})
		newReq := req.WithContext(ctx)
		*req = *newReq

		return 0, nil
	}

	return http.StatusUnauthorized, xerrors.New("no token in request")
}
