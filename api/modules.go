package api

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-messager/types"
	venusTypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs/go-cid"
	"net"
	"net/http"
	"reflect"
	"time"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/venus-auth/cmd/jwtclient"
	"github.com/ipfs-force-community/metrics/ratelimit"
	"go.uber.org/fx"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-messager/api/client"
	"github.com/filecoin-project/venus-messager/api/controller"
	"github.com/filecoin-project/venus-messager/api/jwt"
	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/log"
	"github.com/filecoin-project/venus-messager/service"
)

func RunAPI(lc fx.Lifecycle, jwtCli *jwt.JwtClient, lst net.Listener, log *log.Logger, msgImp *MessageImp, rateLimitCfg *config.RateLimitConfig) error {
	var msgAPI client.Message
	PermissionedProxy(controller.AuthMap, msgImp, &msgAPI.Internal)

	srv := jsonrpc.NewServer()
	if len(rateLimitCfg.Redis) != 0 && jwtCli.Remote != nil && jwtCli.Remote.Cli != nil {
		limiter, err := ratelimit.NewRateLimitHandler(
			rateLimitCfg.Redis,
			nil,
			&jwtclient.ValueFromCtx{},
			jwtclient.WarpLimitFinder(jwtCli.Remote.Cli),
			log,
		)
		if err != nil {
			return err
		}
		var rateLimitAPI client.Message
		limiter.WrapFunctions(&msgAPI, &rateLimitAPI.Internal)
		srv.Register("Message", &rateLimitAPI)
	} else {
		srv.Register("Message", &msgAPI)
	}

	handler := http.NewServeMux()
	handler.Handle("/rpc/v0", srv)
	authMux := jwtclient.NewAuthMux(jwtCli.Local, jwtCli.Remote, handler, log)
	authMux.TrustHandle("/debug/pprof/", http.DefaultServeMux)

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
	Logger              *log.Logger
}

type MessageImp struct {
	AddressSrv *service.AddressService
	MessageSrv *service.MessageService
	NodeSrv    *service.NodeService
	ParamsSrv  *service.SharedParamsService
	log        *log.Logger
}

func (m MessageImp) HasMessageByUid(ctx context.Context, id string) (bool, error) {
	return m.MessageSrv.HasMessageByUid(ctx, id)
}

func (m MessageImp) WaitMessage(ctx context.Context, id string, confidence uint64) (*types.Message, error) {
	return m.MessageSrv.WaitMessage(ctx, id, confidence)
}

func (m MessageImp) ForcePushMessage(ctx context.Context, account string, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (string, error) {
	return m.MessageSrv.PushMessage(ctx, account, msg, meta)
}

func (m MessageImp) PushMessage(ctx context.Context, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (string, error) {
	_, account := ipAccountFromContext(ctx)
	return m.MessageSrv.PushMessage(ctx, account, msg, meta)
}

func (m MessageImp) PushMessageWithId(ctx context.Context, id string, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (string, error) {
	_, account := ipAccountFromContext(ctx)
	return m.MessageSrv.PushMessageWithId(ctx, account, id, msg, meta)
}

func (m MessageImp) GetMessageByUid(ctx context.Context, id string) (*types.Message, error) {
	return m.MessageSrv.GetMessageByUid(ctx, id)
}

func (m MessageImp) GetMessageByCid(ctx context.Context, id cid.Cid) (*types.Message, error) {
	return m.MessageSrv.GetMessageByCid(ctx, id)
}

func (m MessageImp) GetMessageBySignedCid(ctx context.Context, cid cid.Cid) (*types.Message, error) {
	return m.MessageSrv.GetMessageBySignedCid(ctx, cid)
}

func (m MessageImp) GetMessageByUnsignedCid(ctx context.Context, cid cid.Cid) (*types.Message, error) {
	return m.MessageSrv.GetMessageByUnsignedCid(ctx, cid)
}

func (m MessageImp) GetMessageByFromAndNonce(ctx context.Context, from address.Address, nonce uint64) (*types.Message, error) {
	return m.MessageSrv.GetMessageByFromAndNonce(ctx, from, nonce)
}

func (m MessageImp) ListMessage(ctx context.Context) ([]*types.Message, error) {
	return m.MessageSrv.ListMessage(ctx)
}

func (m MessageImp) ListMessageByFromState(ctx context.Context, from address.Address, state types.MessageState, pageIndex, pageSize int) ([]*types.Message, error) {
	return m.MessageSrv.ListMessageByFromState(ctx, from, state, pageIndex, pageSize)
}

func (m MessageImp) ListMessageByAddress(ctx context.Context, addr address.Address) ([]*types.Message, error) {
	return m.MessageSrv.ListMessageByAddress(ctx, addr)
}

func (m MessageImp) ListFailedMessage(ctx context.Context) ([]*types.Message, error) {
	return m.MessageSrv.ListFailedMessage(ctx)
}

func (m MessageImp) ListBlockedMessage(ctx context.Context, addr address.Address, d time.Duration) ([]*types.Message, error) {
	return m.MessageSrv.ListBlockedMessage(ctx, addr, d)
}

func (m MessageImp) UpdateMessageStateByID(ctx context.Context, id string, state types.MessageState) (string, error) {
	return m.MessageSrv.UpdateMessageStateByID(ctx, id, state)
}

func (m MessageImp) UpdateAllFilledMessage(ctx context.Context) (int, error) {
	return m.MessageSrv.UpdateAllFilledMessage(ctx)
}

func (m MessageImp) UpdateFilledMessageByID(ctx context.Context, id string) (string, error) {
	return m.MessageSrv.UpdateFilledMessageByID(ctx, id)
}

func (m MessageImp) ReplaceMessage(ctx context.Context, id string, auto bool, maxFee string, gasLimit int64, gasPremium string, gasFeecap string) (cid.Cid, error) {
	return m.MessageSrv.ReplaceMessage(ctx, id, auto, maxFee, gasLimit, gasPremium, gasFeecap)
}

func (m MessageImp) RepublishMessage(ctx context.Context, id string) (struct{}, error) {
	return m.MessageSrv.RepublishMessage(ctx, id)
}

func (m MessageImp) MarkBadMessage(ctx context.Context, id string) (struct{}, error) {
	return m.MessageSrv.MarkBadMessage(ctx, id)
}

func (m MessageImp) RecoverFailedMsg(ctx context.Context, addr address.Address) ([]string, error) {
	return m.MessageSrv.RecoverFailedMsg(ctx, addr)
}

func (m MessageImp) SaveAddress(ctx context.Context, addr *types.Address) (types.UUID, error) {
	return m.AddressSrv.SaveAddress(ctx, addr)
}

func (m MessageImp) GetAddress(ctx context.Context, addr address.Address) (*types.Address, error) {
	return m.AddressSrv.GetAddress(ctx, addr)
}

func (m MessageImp) HasAddress(ctx context.Context, addr address.Address) (bool, error) {
	return m.AddressSrv.HasAddress(ctx, addr)
}

func (m MessageImp) WalletHas(ctx context.Context, addr address.Address) (bool, error) {
	_, account := ipAccountFromContext(ctx)
	return m.AddressSrv.WalletHas(ctx, account, addr)
}

func (m MessageImp) ListAddress(ctx context.Context) ([]*types.Address, error) {
	return m.AddressSrv.ListAddress(ctx)
}

func (m MessageImp) UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) (address.Address, error) {
	return m.AddressSrv.UpdateNonce(ctx, addr, nonce)
}

func (m MessageImp) DeleteAddress(ctx context.Context, addr address.Address) (address.Address, error) {
	return m.AddressSrv.DeleteAddress(ctx, addr)
}

func (m MessageImp) ForbiddenAddress(ctx context.Context, addr address.Address) (address.Address, error) {
	return m.AddressSrv.ForbiddenAddress(ctx, addr)
}

func (m MessageImp) ActiveAddress(ctx context.Context, addr address.Address) (address.Address, error) {
	return m.AddressSrv.ActiveAddress(ctx, addr)
}

func (m MessageImp) SetSelectMsgNum(ctx context.Context, addr address.Address, num uint64) (address.Address, error) {
	return m.AddressSrv.SetSelectMsgNum(ctx, addr, num)
}

func (m MessageImp) SetFeeParams(ctx context.Context, addr address.Address, gasOverEstimation float64, maxFee, maxFeeCap string) (address.Address, error) {
	return m.AddressSrv.SetFeeParams(ctx, addr, gasOverEstimation, maxFee, maxFeeCap)
}

func (m MessageImp) ClearUnFillMessage(ctx context.Context, addr address.Address) (int, error) {
	return m.MessageSrv.ClearUnFillMessage(ctx, addr)
}

func (m MessageImp) GetSharedParams(ctx context.Context) (*types.SharedParams, error) {
	return m.ParamsSrv.GetSharedParams(ctx)
}

func (m MessageImp) SetSharedParams(ctx context.Context, params *types.SharedParams) (struct{}, error) {
	return m.ParamsSrv.SetSharedParams(ctx, params)
}

func (m MessageImp) RefreshSharedParams(ctx context.Context) (struct{}, error) {
	return m.ParamsSrv.RefreshSharedParams(ctx)
}

func (m MessageImp) SaveNode(ctx context.Context, node *types.Node) (struct{}, error) {
	return m.NodeSrv.SaveNode(ctx, node)
}

func (m MessageImp) GetNode(ctx context.Context, name string) (*types.Node, error) {
	return m.NodeSrv.GetNode(ctx, name)
}

func (m MessageImp) HasNode(ctx context.Context, name string) (bool, error) {
	return m.NodeSrv.HasNode(ctx, name)
}

func (m MessageImp) ListNode(ctx context.Context) ([]*types.Node, error) {
	return m.NodeSrv.ListNode(ctx)
}

func (m MessageImp) DeleteNode(ctx context.Context, name string) (struct{}, error) {
	return m.DeleteNode(ctx, name)
}

func (m MessageImp) SetLogLevel(ctx context.Context, level string) error {
	return m.log.SetLogLevel(ctx, level)
}

func (m MessageImp) Send(ctx context.Context, params types.SendParams) (string, error) {
	return m.MessageSrv.Send(ctx, params)
}

func ipAccountFromContext(ctx context.Context) (string, string) {
	ip, _ := jwtclient.CtxGetTokenLocation(ctx)
	account, _ := jwtclient.CtxGetName(ctx)

	return ip, account
}

var _ client.IMessager = (*MessageImp)(nil)

func NewMessageImp(implParams ImplParams) *MessageImp {
	return &MessageImp{
		AddressSrv: implParams.AddressService,
		MessageSrv: implParams.MessageService,
		NodeSrv:    implParams.NodeService,
		ParamsSrv:  implParams.SharedParamsService,
		log:        implParams.Logger,
	}
}

var AllPermissions = []auth.Permission{"read", "write", "sign", "admin"}
var defaultPerms = []auth.Permission{"read"}

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
			if auth.HasPerm(ctx, defaultPerms, requiredPerm) {
				return fn.Call(args)
			}

			err := xerrors.Errorf("missing permission to invoke '%s', need '%s'", field.Name, requiredPerm)
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
