package api

import (
	"context"
	"time"

	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/fx"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-messager/publisher/pubsub"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"

	"github.com/filecoin-project/venus-messager/service"
	"github.com/filecoin-project/venus-messager/version"
)

type ImplParams struct {
	fx.In
	AddressService      *service.AddressService
	MessageService      *service.MessageService
	NodeService         service.INodeService
	SharedParamsService *service.SharedParamsService
	Net                 pubsub.INet
}

func NewMessageImp(implParams ImplParams) *MessageImp {
	return &MessageImp{
		AddressSrv: implParams.AddressService,
		MessageSrv: implParams.MessageService,
		NodeSrv:    implParams.NodeService,
		ParamsSrv:  implParams.SharedParamsService,
		Net:        implParams.Net,
	}
}

type MessageImp struct {
	AddressSrv *service.AddressService
	MessageSrv *service.MessageService
	NodeSrv    service.INodeService
	ParamsSrv  *service.SharedParamsService
	Net        pubsub.INet
}

func (m MessageImp) HasMessageByUid(ctx context.Context, id string) (bool, error) {
	return m.MessageSrv.HasMessageByUid(ctx, id)
}

func (m MessageImp) WaitMessage(ctx context.Context, id string, confidence uint64) (*types.Message, error) {
	return m.MessageSrv.WaitMessage(ctx, id, confidence)
}

func (m MessageImp) PushMessage(ctx context.Context, msg *venusTypes.Message, meta *types.SendSpec) (string, error) {
	return m.MessageSrv.PushMessage(ctx, msg, meta)
}

func (m MessageImp) PushMessageWithId(ctx context.Context, id string, msg *venusTypes.Message, meta *types.SendSpec) (string, error) {
	return m.MessageSrv.PushMessageWithId(ctx, id, msg, meta)
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

func (m MessageImp) ReplaceMessage(ctx context.Context, params *types.ReplacMessageParams) (cid.Cid, error) {
	return m.MessageSrv.ReplaceMessage(ctx, params)
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
	return m.AddressSrv.WalletHas(ctx, addr)
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

func (m MessageImp) SetFeeParams(ctx context.Context, params *types.AddressSpec) error {
	return m.AddressSrv.SetFeeParams(ctx, params)
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

func (m MessageImp) Send(ctx context.Context, params types.QuickSendParams) (string, error) {
	return m.MessageSrv.Send(ctx, params)
}

func (m MessageImp) NetFindPeer(ctx context.Context, peerID peer.ID) (peer.AddrInfo, error) {
	return m.Net.FindPeer(ctx, peerID)
}

func (m MessageImp) NetConnect(ctx context.Context, pi peer.AddrInfo) error {
	return m.Net.Connect(ctx, pi)
}

func (m MessageImp) NetPeers(ctx context.Context) ([]peer.AddrInfo, error) {
	return m.Net.Peers(ctx)
}

func (m MessageImp) NetAddrsListen(ctx context.Context) (peer.AddrInfo, error) {
	return m.Net.AddrListen(ctx)
}

var _ messager.IMessager = (*MessageImp)(nil)

func (m MessageImp) Version(_ context.Context) (venusTypes.Version, error) {
	return venusTypes.Version{
		Version: version.Version,
	}, nil
}

func (m MessageImp) SetLogLevel(ctx context.Context, subSystem, level string) error {
	return logging.SetLogLevel(subSystem, level)
}

func (m MessageImp) LogList(ctx context.Context) ([]string, error) {
	return logging.GetSubsystems(), nil
}
