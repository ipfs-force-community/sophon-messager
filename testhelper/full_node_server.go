package testhelper

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/filecoin-project/venus/venus-shared/api/chain/v0/mock"
	"github.com/golang/mock/gomock"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

type FullNodeServer struct {
	Stop func(ctx context.Context) error

	Port  string
	Token string
}

func MockFullNodeServer(t *testing.T) (*FullNodeServer, error) {
	mAddr, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		return nil, err
	}
	apiListener, err := manet.Listen(mAddr)
	if err != nil {
		return nil, err
	}
	lst := manet.NetListener(apiListener)

	ctrl := gomock.NewController(t)
	full := mock.NewMockFullNode(ctrl)

	srv := jsonrpc.NewServer()
	srv.Register("Filecoin", full)
	handler := http.NewServeMux()
	handler.Handle("/rpc/v0", srv)
	handler.Handle("/rpc/v1", srv)

	localAuthCli, token, err := jwtclient.NewLocalAuthClient()
	if err != nil {
		return nil, fmt.Errorf("failed to generate local auth client %v", err)
	}

	authMux := jwtclient.NewAuthMux(localAuthCli, nil, handler)
	apiserv := &http.Server{
		Handler: authMux,
	}

	go func() {
		t.Logf("start node rpcserver: %v", lst.Addr())
		if err := apiserv.Serve(lst); err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Errorf("start node rpcserver failed: %v", err)
		}
	}()

	return &FullNodeServer{
		Stop: func(ctx context.Context) error {
			return apiserv.Shutdown(ctx)
		},
		Port:  strings.Split(lst.Addr().String(), ":")[1],
		Token: string(token),
	}, nil
}
