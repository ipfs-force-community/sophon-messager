package api

import (
	"context"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-auth/cmd/jwtclient"
	"github.com/filecoin-project/venus-messager/log"
	"github.com/filecoin-project/venus-messager/service"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/ipfs/go-cid"
	"go.uber.org/fx"
)

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

func (m MessageImp) ForcePushMessage(ctx context.Context, account string, msg *venusTypes.Message, meta *types.SendSpec) (string, error) {
	return m.MessageSrv.PushMessage(ctx, account, msg, meta)
}

func (m MessageImp) ForcePushMessageWithId(ctx context.Context, account string, id string, msg *venusTypes.Message, meta *types.SendSpec) (string, error) {
	return m.MessageSrv.PushMessageWithId(ctx, account, id, msg, meta)
}

func (m MessageImp) PushMessage(ctx context.Context, msg *venusTypes.Message, meta *types.SendSpec) (string, error) {
	_, account := ipAccountFromContext(ctx)
	return m.MessageSrv.PushMessage(ctx, account, msg, meta)
}

func (m MessageImp) PushMessageWithId(ctx context.Context, id string, msg *venusTypes.Message, meta *types.SendSpec) (string, error) {
	_, account := ipAccountFromContext(ctx)
	return m.MessageSrv.PushMessageWithId(ctx, account, id, msg, meta)
}

func (m MessageImp) GetMessageByUid(ctx context.Context, id string) (*types.Message, error) {
	return m.MessageSrv.GetMessageByUid(ctx, id)
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

func (m MessageImp) ListMessageByFromState(ctx context.Context, from address.Address, state types.MessageState, isAsc bool, pageIndex, pageSize int) ([]*types.Message, error) {
	return m.MessageSrv.ListMessageByFromState(ctx, from, state, isAsc, pageIndex, pageSize)
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

func (m MessageImp) UpdateMessageStateByID(ctx context.Context, id string, state types.MessageState) error {
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

func (m MessageImp) RepublishMessage(ctx context.Context, id string) error {
	return m.MessageSrv.RepublishMessage(ctx, id)
}

func (m MessageImp) MarkBadMessage(ctx context.Context, id string) error {
	return m.MessageSrv.MarkBadMessage(ctx, id)
}

func (m MessageImp) RecoverFailedMsg(ctx context.Context, addr address.Address) ([]string, error) {
	return m.MessageSrv.RecoverFailedMsg(ctx, addr)
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

func (m MessageImp) UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) error {
	return m.AddressSrv.UpdateNonce(ctx, addr, nonce)
}

func (m MessageImp) DeleteAddress(ctx context.Context, addr address.Address) error {
	return m.AddressSrv.DeleteAddress(ctx, addr)
}

func (m MessageImp) ForbiddenAddress(ctx context.Context, addr address.Address) error {
	return m.AddressSrv.ForbiddenAddress(ctx, addr)
}

func (m MessageImp) ActiveAddress(ctx context.Context, addr address.Address) error {
	return m.AddressSrv.ActiveAddress(ctx, addr)
}

func (m MessageImp) SetSelectMsgNum(ctx context.Context, addr address.Address, num uint64) error {
	return m.AddressSrv.SetSelectMsgNum(ctx, addr, num)
}

func (m MessageImp) SetFeeParams(ctx context.Context, addr address.Address, gasOverEstimation float64, maxFee, maxFeeCap string) error {
	return m.AddressSrv.SetFeeParams(ctx, addr, gasOverEstimation, maxFee, maxFeeCap)
}

func (m MessageImp) ClearUnFillMessage(ctx context.Context, addr address.Address) (int, error) {
	return m.MessageSrv.ClearUnFillMessage(ctx, addr)
}

func (m MessageImp) GetSharedParams(ctx context.Context) (*types.SharedSpec, error) {
	return m.ParamsSrv.GetSharedParams(ctx)
}

func (m MessageImp) SetSharedParams(ctx context.Context, params *types.SharedSpec) error {
	return m.ParamsSrv.SetSharedParams(ctx, params)
}

func (m MessageImp) RefreshSharedParams(ctx context.Context) error {
	return m.ParamsSrv.RefreshSharedParams(ctx)
}

func (m MessageImp) SaveNode(ctx context.Context, node *types.Node) error {
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

func (m MessageImp) DeleteNode(ctx context.Context, name string) error {
	return m.NodeSrv.DeleteNode(ctx, name)
}

func (m MessageImp) SetLogLevel(ctx context.Context, level string) error {
	return m.log.SetLogLevel(ctx, level)
}

func (m MessageImp) Send(ctx context.Context, params types.QuickSendParams) (string, error) {
	return m.MessageSrv.Send(ctx, params)
}

func ipAccountFromContext(ctx context.Context) (string, string) {
	ip, _ := jwtclient.CtxGetTokenLocation(ctx)
	account, _ := jwtclient.CtxGetName(ctx)

	return ip, account
}

var _ messager.IMessager = (*MessageImp)(nil)

func NewMessageImp(implParams ImplParams) *MessageImp {
	return &MessageImp{
		AddressSrv: implParams.AddressService,
		MessageSrv: implParams.MessageService,
		NodeSrv:    implParams.NodeService,
		ParamsSrv:  implParams.SharedParamsService,
		log:        implParams.Logger,
	}
}
