package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"reflect"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/venus-auth/core"
	"go.uber.org/fx"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-messager/api/client"
	"github.com/filecoin-project/venus-messager/api/controller"
	"github.com/filecoin-project/venus-messager/api/jwt"
	"github.com/filecoin-project/venus-messager/gateway"
	"github.com/filecoin-project/venus-messager/log"
	"github.com/filecoin-project/venus-messager/service"
)

func RunAPI(lc fx.Lifecycle, jwtClient jwt.IJwtClient, lst net.Listener, log *log.Logger, msgImp *MessageImp) error {
	var msgAPI client.Message
	PermissionedProxy(controller.AuthMap, msgImp, &msgAPI.Internal)

	srv := jsonrpc.NewServer()
	srv.Register("Message", &msgAPI)

	handler := http.NewServeMux()
	handler.Handle("/rpc/v0", srv)
	authMux := jwt.NewAuthMux(jwtClient, log, handler)
	authMux.TruthHandle("/debug/pprof/", http.DefaultServeMux)

	apiserv := &http.Server{
		Handler: authMux,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Info("Start rpcserver ", lst.Addr())
				if err := apiserv.Serve(lst); err != nil {
					log.Errorf("Start rpcserver failed: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return lst.Close()
		},
	})
	return nil
}

type ImplParams struct {
	fx.In
	AddressService      *service.AddressService
	MessageService      *service.MessageService
	NodeService         *service.NodeService
	SharedParamsService *service.SharedParamsService
	GatewayService      *gateway.GatewayService `optional:"true"`
	Logger              *log.Logger
}

type MessageImp struct {
	*service.AddressService
	*service.MessageService
	*service.NodeService
	*service.SharedParamsService
	*gateway.GatewayService
	*log.Logger
}

var _ client.IMessager = (*MessageImp)(nil)

func NewMessageImp(implParams ImplParams) *MessageImp {
	return &MessageImp{
		AddressService:      implParams.AddressService,
		MessageService:      implParams.MessageService,
		NodeService:         implParams.NodeService,
		SharedParamsService: implParams.SharedParamsService,
		GatewayService:      implParams.GatewayService,
		Logger:              implParams.Logger,
	}
}

func PermissionedProxy(permMap map[string]string, in interface{}, out interface{}) {
	rint := reflect.ValueOf(out).Elem()
	ra := reflect.ValueOf(in)

	for f := 0; f < rint.NumField(); f++ {
		field := rint.Type().Field(f)

		fn := ra.MethodByName(field.Name)
		requiredPerm, ok := permMap[field.Name]
		if !ok {
			panic(fmt.Sprintf("'%s' not found perm", field.Name))
		}

		rint.Field(f).Set(reflect.MakeFunc(field.Type, func(args []reflect.Value) (results []reflect.Value) {
			ctx := args[0].Interface().(context.Context)
			err := hasPerm(ctx, requiredPerm)
			if err == nil {
				return fn.Call(args)
			}

			err = xerrors.Errorf("missing permission to invoke '%s' %s", field.Name, err.Error())
			rerr := reflect.ValueOf(&err).Elem()

			if field.Type.NumOut() == 2 {
				return []reflect.Value{
					reflect.Zero(field.Type.Out(0)),
					rerr,
				}
			} else {
				return []reflect.Value{rerr}
			}
		}))

	}
}

func hasPerm(ctx context.Context, requiredPerm string) error {
	perms, ok := ctx.Value(core.PermCtxKey).([]core.Permission)
	if !ok {
		return xerrors.Errorf("unknown perm type %T", ctx.Value(core.PermCtxKey))
	}

	for _, p := range perms {
		if requiredPerm == p {
			return nil
		}
	}

	return xerrors.Errorf("(need %s) has %v", requiredPerm, perms)
}
