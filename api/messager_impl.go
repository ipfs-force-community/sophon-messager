package api

import (
	"context"
	"fmt"
	"time"

	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/fx"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/venus-messager/publisher/pubsub"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"

	"github.com/filecoin-project/venus-messager/service"
	"github.com/filecoin-project/venus-messager/version"

	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/jwtclient"
)

type ImplParams struct {
	fx.In
	AddressService      *service.AddressService
	MessageService      *service.MessageService
	NodeService         service.INodeService
	SharedParamsService *service.SharedParamsService
	Net                 pubsub.INet
	AuthClient          jwtclient.IAuthClient
	NodeClient          v1.FullNode
}

func NewMessageImp(implParams ImplParams) *MessageImp {
	return &MessageImp{
		AddressSrv: implParams.AddressService,
		MessageSrv: implParams.MessageService,
		NodeSrv:    implParams.NodeService,
		ParamsSrv:  implParams.SharedParamsService,
		Net:        implParams.Net,
		AuthClient: implParams.AuthClient,
		NodeClient: implParams.NodeClient,
	}
}

type MessageImp struct {
	AddressSrv service.IAddressService
	MessageSrv service.IMessageService
	NodeSrv    service.INodeService
	ParamsSrv  *service.SharedParamsService
	Net        pubsub.INet
	AuthClient jwtclient.IAuthClient
	NodeClient v1.FullNode
}

var _ messager.IMessager = (*MessageImp)(nil)

func (m MessageImp) HasMessageByUid(ctx context.Context, id string) (bool, error) {
	return m.MessageSrv.HasMessageByUid(ctx, id)
}

func (m MessageImp) WaitMessage(ctx context.Context, id string, confidence uint64) (*types.Message, error) {
	msg, err := m.MessageSrv.GetMessageByUid(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get message by id error: %w", err)
	}
	if checkErr := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, msg.From); checkErr != nil {
		return nil, checkErr
	}
	return m.MessageSrv.WaitMessage(ctx, id, confidence)
}

func (m MessageImp) PushMessage(ctx context.Context, msg *venusTypes.Message, meta *types.SendSpec) (string, error) {
	id := venusTypes.NewUUID().String()
	return m.MessageSrv.PushMessageWithId(ctx, id, msg, meta)
}

func (m MessageImp) PushMessageWithId(ctx context.Context, id string, msg *venusTypes.Message, meta *types.SendSpec) (string, error) {
	// replace address
	if msg.From.Protocol() == address.ID {
		fromA, err := m.NodeClient.StateAccountKey(ctx, msg.From, venusTypes.EmptyTSK)
		if err != nil {
			return "", fmt.Errorf("getting key address %s failed: %w", msg.From, err)
		}
		log.Warnf("Push from ID address (%s), adjusting to %s", msg.From, fromA)
		msg.From = fromA
	}

	if err := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, msg.From); err != nil {
		return "", err
	}
	return m.MessageSrv.PushMessageWithId(ctx, id, msg, meta)
}

func (m MessageImp) GetMessageByUid(ctx context.Context, id string) (*types.Message, error) {
	msg, err := m.MessageSrv.GetMessageByUid(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get message by id error: %w", err)
	}
	if checkErr := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, msg.From); checkErr != nil {
		return nil, checkErr
	}
	return msg, nil
}

func (m MessageImp) GetMessageBySignedCid(ctx context.Context, cid cid.Cid) (*types.Message, error) {
	msg, err := m.MessageSrv.GetMessageBySignedCid(ctx, cid)
	if err != nil {
		return nil, fmt.Errorf("get message by id error: %w", err)
	}
	if checkErr := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, msg.From); checkErr != nil {
		return nil, checkErr
	}
	return msg, nil
}

func (m MessageImp) GetMessageByUnsignedCid(ctx context.Context, cid cid.Cid) (*types.Message, error) {
	msg, err := m.MessageSrv.GetMessageByUnsignedCid(ctx, cid)
	if err != nil {
		return nil, fmt.Errorf("get message by id error: %w", err)
	}
	if checkErr := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, msg.From); checkErr != nil {
		return nil, checkErr
	}
	return msg, nil
}

func (m MessageImp) GetMessageByFromAndNonce(ctx context.Context, from address.Address, nonce uint64) (*types.Message, error) {
	msg, err := m.MessageSrv.GetMessageByFromAndNonce(ctx, from, nonce)
	if err != nil {
		return nil, fmt.Errorf("get message by id error: %w", err)
	}
	if checkErr := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, msg.From); checkErr != nil {
		return nil, checkErr
	}
	return msg, nil
}

func (m MessageImp) ListMessage(ctx context.Context, p *types.MsgQueryParams) ([]*types.Message, error) {
	// only admin can list all message
	if len(p.From) == 0 {
		if !isAdmin(ctx) {
			signers, err := getSigners(ctx, m.AuthClient)
			if err != nil {
				return nil, err
			}
			if len(signers) == 0 {
				return nil, nil
			}
			p.From = signers
		}
	} else {
		if err := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, p.From...); err != nil {
			return nil, err
		}
	}
	return m.MessageSrv.ListMessage(ctx, p)
}

func (m MessageImp) ListMessageByFromState(ctx context.Context, from address.Address, state types.MessageState, isAsc bool, pageIndex, pageSize int) ([]*types.Message, error) {
	return m.MessageSrv.ListMessageByFromState(ctx, from, state, isAsc, pageIndex, pageSize)
}

func (m MessageImp) ListMessageByAddress(ctx context.Context, addr address.Address) ([]*types.Message, error) {
	return m.MessageSrv.ListMessageByAddress(ctx, addr)
}

func (m MessageImp) ListFailedMessage(ctx context.Context) ([]*types.Message, error) {
	if isAdmin(ctx) {
		return m.MessageSrv.ListFailedMessage(ctx, &types.MsgQueryParams{})
	}
	signers, err := getSigners(ctx, m.AuthClient)
	if err != nil {
		return nil, err
	}
	if len(signers) == 0 {
		return nil, nil
	}
	return m.MessageSrv.ListFailedMessage(ctx, &types.MsgQueryParams{From: signers})
}

func (m MessageImp) ListBlockedMessage(ctx context.Context, addr address.Address, d time.Duration) ([]*types.Message, error) {
	if err := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, addr); err != nil {
		return nil, err
	}
	return m.MessageSrv.ListBlockedMessage(ctx, &types.MsgQueryParams{From: []address.Address{addr}}, d)
}

func (m MessageImp) UpdateMessageStateByID(ctx context.Context, id string, state types.MessageState) error {
	msg, err := m.MessageSrv.GetMessageByUid(ctx, id)
	if err != nil {
		return fmt.Errorf("get message by id error: %w", err)
	}
	if checkErr := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, msg.From); checkErr != nil {
		return checkErr
	}
	return m.MessageSrv.UpdateMessageStateByID(ctx, id, state)
}

func (m MessageImp) UpdateAllFilledMessage(ctx context.Context) (int, error) {
	return m.MessageSrv.UpdateAllFilledMessage(ctx)
}

func (m MessageImp) UpdateFilledMessageByID(ctx context.Context, id string) (string, error) {
	msg, err := m.MessageSrv.GetMessageByUid(ctx, id)
	if err != nil {
		return "", fmt.Errorf("get message by id error: %w", err)
	}
	if checkErr := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, msg.From); checkErr != nil {
		return "", checkErr
	}
	if msg.State == types.OnChainMsg || msg.State == types.NonceConflictMsg {
		return "", fmt.Errorf("message state(%s) has been final, can not update", msg.State)
	}
	return m.MessageSrv.UpdateFilledMessageByID(ctx, id)
}

func (m MessageImp) ReplaceMessage(ctx context.Context, params *types.ReplacMessageParams) (cid.Cid, error) {
	msg, err := m.MessageSrv.GetMessageByUid(ctx, params.ID)
	if err != nil {
		return cid.Undef, fmt.Errorf("get message by id error: %w", err)
	}
	if checkErr := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, msg.From); checkErr != nil {
		return cid.Undef, checkErr
	}
	return m.MessageSrv.ReplaceMessage(ctx, params)
}

func (m MessageImp) RepublishMessage(ctx context.Context, id string) error {
	return m.MessageSrv.RepublishMessage(ctx, id)
}

func (m MessageImp) MarkBadMessage(ctx context.Context, id string) error {
	msg, err := m.MessageSrv.GetMessageByUid(ctx, id)
	if err != nil {
		return fmt.Errorf("get message by id error: %w", err)
	}
	if checkErr := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, msg.From); checkErr != nil {
		return checkErr
	}
	return m.MessageSrv.MarkBadMessage(ctx, id)
}

func (m MessageImp) RecoverFailedMsg(ctx context.Context, addr address.Address) ([]string, error) {
	if err := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, addr); err != nil {
		return nil, err
	}
	return m.MessageSrv.RecoverFailedMsg(ctx, addr)
}

func (m MessageImp) GetAddress(ctx context.Context, addr address.Address) (*types.Address, error) {
	if err := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, addr); err != nil {
		return nil, err
	}
	return m.AddressSrv.GetAddress(ctx, addr)
}

func (m MessageImp) HasAddress(ctx context.Context, addr address.Address) (bool, error) {
	if err := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, addr); err != nil {
		return false, err
	}
	return m.AddressSrv.HasAddress(ctx, addr)
}

func (m MessageImp) WalletHas(ctx context.Context, addr address.Address) (bool, error) {
	if err := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, addr); err != nil {
		return false, err
	}
	return m.AddressSrv.WalletHas(ctx, addr)
}

func (m MessageImp) ListAddress(ctx context.Context) ([]*types.Address, error) {
	msgs, err := m.AddressSrv.ListAddress(ctx)
	if err != nil {
		return nil, err
	}
	var result []*types.Address
	for _, msg := range msgs {
		if err := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, msg.Addr); err == nil {
			result = append(result, msg)
		}
	}
	return result, nil
}

func (m MessageImp) UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) error {
	if err := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, addr); err != nil {
		return err
	}
	return m.AddressSrv.UpdateNonce(ctx, addr, nonce)
}

func (m MessageImp) DeleteAddress(ctx context.Context, addr address.Address) error {
	if err := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, addr); err != nil {
		return err
	}
	return m.AddressSrv.DeleteAddress(ctx, addr)
}

func (m MessageImp) ForbiddenAddress(ctx context.Context, addr address.Address) error {
	if err := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, addr); err != nil {
		return err
	}
	return m.AddressSrv.ForbiddenAddress(ctx, addr)
}

func (m MessageImp) ActiveAddress(ctx context.Context, addr address.Address) error {
	if err := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, addr); err != nil {
		return err
	}
	return m.AddressSrv.ActiveAddress(ctx, addr)
}

func (m MessageImp) SetSelectMsgNum(ctx context.Context, addr address.Address, num uint64) error {
	if err := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, addr); err != nil {
		return err
	}
	return m.AddressSrv.SetSelectMsgNum(ctx, addr, num)
}

func (m MessageImp) SetFeeParams(ctx context.Context, params *types.AddressSpec) error {
	if err := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, params.Address); err != nil {
		return err
	}
	return m.AddressSrv.SetFeeParams(ctx, params)
}

func (m MessageImp) ClearUnFillMessage(ctx context.Context, addr address.Address) (int, error) {
	if err := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, addr); err != nil {
		return 0, err
	}
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
	if err := jwtclient.CheckPermissionBySigner(ctx, m.AuthClient, params.From); err != nil {
		return "", err
	}
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

// isAdmin check if the user is admin and return signers of the user
func isAdmin(ctx context.Context) bool {
	return auth.HasPerm(ctx, nil, core.PermAdmin)
}

func getSigners(ctx context.Context, client jwtclient.IAuthClient) ([]address.Address, error) {
	signers := []address.Address{}

	user, exit := jwtclient.CtxGetName(ctx)
	if !exit {
		return nil, jwtclient.ErrorUserNotFound
	}
	resp, err := client.ListSigners(ctx, user)
	if err != nil {
		return nil, err
	}
	for _, res := range resp {
		signers = append(signers, res.Signer)
	}

	return signers, nil
}
