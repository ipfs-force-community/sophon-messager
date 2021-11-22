package client

import (
	"context"
	"time"

	"github.com/filecoin-project/go-address"
	venusTypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/venus-messager/types"
)

type IMessager interface {
	HasMessageByUid(ctx context.Context, id string) (bool, error)                                                                                  //perm:read
	WaitMessage(ctx context.Context, id string, confidence uint64) (*types.Message, error)                                                         //perm:read
	ForcePushMessage(ctx context.Context, account string, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (string, error)                    //perm:admin
	PushMessage(ctx context.Context, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (string, error)                                         //perm:write
	PushMessageWithId(ctx context.Context, id string, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (string, error)                        //perm:write
	GetMessageByUid(ctx context.Context, id string) (*types.Message, error)                                                                        //perm:read
	GetMessageByCid(ctx context.Context, id cid.Cid) (*types.Message, error)                                                                       //perm:read
	GetMessageBySignedCid(ctx context.Context, cid cid.Cid) (*types.Message, error)                                                                //perm:read
	GetMessageByUnsignedCid(ctx context.Context, cid cid.Cid) (*types.Message, error)                                                              //perm:read
	GetMessageByFromAndNonce(ctx context.Context, from address.Address, nonce uint64) (*types.Message, error)                                      //perm:read
	ListMessage(ctx context.Context) ([]*types.Message, error)                                                                                     //perm:admin
	ListMessageByFromState(ctx context.Context, from address.Address, state types.MessageState, pageIndex, pageSize int) ([]*types.Message, error) //perm:admin
	ListMessageByAddress(ctx context.Context, addr address.Address) ([]*types.Message, error)                                                      //perm:admin
	ListFailedMessage(ctx context.Context) ([]*types.Message, error)                                                                               //perm:admin
	ListBlockedMessage(ctx context.Context, addr address.Address, d time.Duration) ([]*types.Message, error)                                       //perm:admin
	UpdateMessageStateByID(ctx context.Context, id string, state types.MessageState) (string, error)                                               //perm:admin
	UpdateAllFilledMessage(ctx context.Context) (int, error)                                                                                       //perm:admin
	UpdateFilledMessageByID(ctx context.Context, id string) (string, error)                                                                        //perm:admin
	ReplaceMessage(ctx context.Context, id string, auto bool, maxFee string, gasLimit int64, gasPremium string, gasFeecap string) (cid.Cid, error) //perm:admin
	RepublishMessage(ctx context.Context, id string) (struct{}, error)                                                                             //perm:admin
	MarkBadMessage(ctx context.Context, id string) (struct{}, error)                                                                               //perm:admin
	RecoverFailedMsg(ctx context.Context, addr address.Address) ([]string, error)                                                                  //perm:admin

	SaveAddress(ctx context.Context, address *types.Address) (types.UUID, error)                                                          //perm:admin
	GetAddress(ctx context.Context, addr address.Address) (*types.Address, error)                                                         //perm:admin
	HasAddress(ctx context.Context, addr address.Address) (bool, error)                                                                   //perm:read
	WalletHas(ctx context.Context, addr address.Address) (bool, error)                                                                    //perm:read
	ListAddress(ctx context.Context) ([]*types.Address, error)                                                                            //perm:admin
	UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) (address.Address, error)                                         //perm:admin
	DeleteAddress(ctx context.Context, addr address.Address) (address.Address, error)                                                     //perm:admin
	ForbiddenAddress(ctx context.Context, addr address.Address) (address.Address, error)                                                  //perm:admin
	ActiveAddress(ctx context.Context, addr address.Address) (address.Address, error)                                                     //perm:admin
	SetSelectMsgNum(ctx context.Context, addr address.Address, num uint64) (address.Address, error)                                       //perm:admin
	SetFeeParams(ctx context.Context, addr address.Address, gasOverEstimation float64, maxFee, maxFeeCap string) (address.Address, error) //perm:admin
	ClearUnFillMessage(ctx context.Context, addr address.Address) (int, error)                                                            //perm:admin

	GetSharedParams(ctx context.Context) (*types.SharedParams, error)                  //perm:admin
	SetSharedParams(ctx context.Context, params *types.SharedParams) (struct{}, error) //perm:admin
	RefreshSharedParams(ctx context.Context) (struct{}, error)                         //perm:admin

	SaveNode(ctx context.Context, node *types.Node) (struct{}, error) //perm:admin
	GetNode(ctx context.Context, name string) (*types.Node, error)    //perm:admin
	HasNode(ctx context.Context, name string) (bool, error)           //perm:admin
	ListNode(ctx context.Context) ([]*types.Node, error)              //perm:admin
	DeleteNode(ctx context.Context, name string) (struct{}, error)    //perm:admin

	SetLogLevel(ctx context.Context, level string) error //perm:admin

	Send(ctx context.Context, params types.SendParams) (string, error) //perm:admin
}

var _ IMessager = (*Message)(nil)

type Message struct {
	Internal struct {
		HasMessageByUid          func(ctx context.Context, id string) (bool, error)
		WaitMessage              func(ctx context.Context, id string, confidence uint64) (*types.Message, error)
		ForcePushMessage         func(ctx context.Context, account string, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (string, error)
		PushMessage              func(ctx context.Context, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (string, error)
		PushMessageWithId        func(ctx context.Context, id string, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (string, error)
		GetMessageByUid          func(ctx context.Context, id string) (*types.Message, error)
		GetMessageByCid          func(ctx context.Context, id cid.Cid) (*types.Message, error)
		GetMessageBySignedCid    func(ctx context.Context, cid cid.Cid) (*types.Message, error)
		GetMessageByUnsignedCid  func(ctx context.Context, cid cid.Cid) (*types.Message, error)
		GetMessageByFromAndNonce func(ctx context.Context, from address.Address, nonce uint64) (*types.Message, error)
		ListMessage              func(ctx context.Context) ([]*types.Message, error)
		ListMessageByAddress     func(ctx context.Context, addr address.Address) ([]*types.Message, error)
		ListMessageByFromState   func(ctx context.Context, from address.Address, state types.MessageState, pageIndex, pageSize int) ([]*types.Message, error)
		ListFailedMessage        func(ctx context.Context) ([]*types.Message, error)
		ListBlockedMessage       func(ctx context.Context, addr address.Address, d time.Duration) ([]*types.Message, error)
		UpdateMessageStateByID   func(ctx context.Context, id string, state types.MessageState) (string, error)
		UpdateAllFilledMessage   func(ctx context.Context) (int, error)
		UpdateFilledMessageByID  func(ctx context.Context, id string) (string, error)
		ReplaceMessage           func(ctx context.Context, id string, auto bool, maxFee string, gasLimit int64, gasPremium string, gasFeecap string) (cid.Cid, error)
		RepublishMessage         func(ctx context.Context, id string) (struct{}, error)
		MarkBadMessage           func(ctx context.Context, id string) (struct{}, error)
		RecoverFailedMsg         func(ctx context.Context, addr address.Address) ([]string, error)

		SaveAddress        func(ctx context.Context, address *types.Address) (types.UUID, error)
		GetAddress         func(ctx context.Context, addr address.Address) (*types.Address, error)
		HasAddress         func(ctx context.Context, addr address.Address) (bool, error)
		WalletHas          func(ctx context.Context, addr address.Address) (bool, error)
		ListAddress        func(ctx context.Context) ([]*types.Address, error)
		UpdateNonce        func(ctx context.Context, addr address.Address, nonce uint64) (address.Address, error)
		DeleteAddress      func(ctx context.Context, addr address.Address) (address.Address, error)
		ForbiddenAddress   func(ctx context.Context, addr address.Address) (address.Address, error)
		ActiveAddress      func(ctx context.Context, addr address.Address) (address.Address, error)
		SetSelectMsgNum    func(ctx context.Context, addr address.Address, num uint64) (address.Address, error)
		SetFeeParams       func(ctx context.Context, addr address.Address, gasOverEstimation float64, maxFee, maxFeeCap string) (address.Address, error)
		ClearUnFillMessage func(ctx context.Context, addr address.Address) (int, error)

		GetSharedParams     func(context.Context) (*types.SharedParams, error)
		SetSharedParams     func(context.Context, *types.SharedParams) (struct{}, error)
		RefreshSharedParams func(ctx context.Context) (struct{}, error)

		SaveNode   func(ctx context.Context, node *types.Node) (struct{}, error)
		GetNode    func(ctx context.Context, name string) (*types.Node, error)
		HasNode    func(ctx context.Context, name string) (bool, error)
		ListNode   func(ctx context.Context) ([]*types.Node, error)
		DeleteNode func(ctx context.Context, name string) (struct{}, error)

		SetLogLevel func(ctx context.Context, level string) error

		Send func(ctx context.Context, params types.SendParams) (string, error)
	}
}

func (message *Message) ForcePushMessage(ctx context.Context, account string, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (string, error) {
	return message.ForcePushMessage(ctx, account, msg, meta)
}

func (message *Message) HasMessageByUid(ctx context.Context, id string) (bool, error) {
	return message.Internal.HasMessageByUid(ctx, id)
}

func (message *Message) PushMessage(ctx context.Context, msg *venusTypes.Message, meta *types.MsgMeta) (string, error) {
	return message.Internal.PushMessage(ctx, msg, meta)
}

func (message *Message) PushMessageWithId(ctx context.Context, id string, msg *venusTypes.UnsignedMessage, meta *types.MsgMeta) (string, error) {
	return message.Internal.PushMessageWithId(ctx, id, msg, meta)
}

func (message *Message) GetMessageByUid(ctx context.Context, id string) (*types.Message, error) {
	return message.Internal.GetMessageByUid(ctx, id)
}

func (message *Message) GetMessageByCid(ctx context.Context, id cid.Cid) (*types.Message, error) {
	return message.Internal.GetMessageByCid(ctx, id)
}

func (message *Message) GetMessageByUnsignedCid(ctx context.Context, cid cid.Cid) (*types.Message, error) {
	return message.Internal.GetMessageByUnsignedCid(ctx, cid)
}

func (message *Message) GetMessageBySignedCid(ctx context.Context, cid cid.Cid) (*types.Message, error) {
	return message.Internal.GetMessageBySignedCid(ctx, cid)
}

func (message *Message) GetMessageByFromAndNonce(ctx context.Context, from address.Address, nonce uint64) (*types.Message, error) {
	return message.Internal.GetMessageByFromAndNonce(ctx, from, nonce)
}

func (message *Message) ListMessage(ctx context.Context) ([]*types.Message, error) {
	return message.Internal.ListMessage(ctx)
}

func (message *Message) ListMessageByFromState(ctx context.Context, from address.Address, state types.MessageState, pageIndex, pageSize int) ([]*types.Message, error) {
	return message.Internal.ListMessageByFromState(ctx, from, state, pageIndex, pageSize)
}

func (message *Message) ListMessageByAddress(ctx context.Context, addr address.Address) ([]*types.Message, error) {
	return message.Internal.ListMessageByAddress(ctx, addr)
}

func (message *Message) ListFailedMessage(ctx context.Context) ([]*types.Message, error) {
	return message.Internal.ListFailedMessage(ctx)
}

func (message *Message) ListBlockedMessage(ctx context.Context, addr address.Address, d time.Duration) ([]*types.Message, error) {
	return message.Internal.ListBlockedMessage(ctx, addr, d)
}

func (message *Message) UpdateMessageStateByID(ctx context.Context, id string, state types.MessageState) (string, error) {
	return message.Internal.UpdateMessageStateByID(ctx, id, state)
}

func (message *Message) UpdateAllFilledMessage(ctx context.Context) (int, error) {
	return message.Internal.UpdateAllFilledMessage(ctx)
}

func (message *Message) UpdateFilledMessageByID(ctx context.Context, id string) (string, error) {
	return message.Internal.UpdateFilledMessageByID(ctx, id)
}

func (message *Message) ReplaceMessage(ctx context.Context, id string, auto bool, maxFee string, gasLimit int64, gasPremium string, gasFeecap string) (cid.Cid, error) {
	return message.Internal.ReplaceMessage(ctx, id, auto, maxFee, gasLimit, gasPremium, gasFeecap)
}

func (message *Message) RepublishMessage(ctx context.Context, id string) (struct{}, error) {
	return message.Internal.RepublishMessage(ctx, id)
}

func (message *Message) MarkBadMessage(ctx context.Context, id string) (struct{}, error) {
	return message.Internal.MarkBadMessage(ctx, id)
}

func (message *Message) WaitMessage(ctx context.Context, id string, confidence uint64) (*types.Message, error) {
	return message.Internal.WaitMessage(ctx, id, confidence)
}

func (message *Message) RecoverFailedMsg(ctx context.Context, addr address.Address) ([]string, error) {
	return message.Internal.RecoverFailedMsg(ctx, addr)
}

///////  address ///////

func (message *Message) SaveAddress(ctx context.Context, address *types.Address) (types.UUID, error) {
	return message.Internal.SaveAddress(ctx, address)
}

func (message *Message) GetAddress(ctx context.Context, addr address.Address) (*types.Address, error) {
	return message.Internal.GetAddress(ctx, addr)
}

func (message *Message) HasAddress(ctx context.Context, addr address.Address) (bool, error) {
	return message.Internal.HasAddress(ctx, addr)
}

func (message *Message) WalletHas(ctx context.Context, addr address.Address) (bool, error) {
	return message.Internal.WalletHas(ctx, addr)
}

func (message *Message) ListAddress(ctx context.Context) ([]*types.Address, error) {
	return message.Internal.ListAddress(ctx)
}

func (message *Message) UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) (address.Address, error) {
	return message.Internal.UpdateNonce(ctx, addr, nonce)
}

func (message *Message) DeleteAddress(ctx context.Context, addr address.Address) (address.Address, error) {
	return message.Internal.DeleteAddress(ctx, addr)
}

func (message *Message) ForbiddenAddress(ctx context.Context, addr address.Address) (address.Address, error) {
	return message.Internal.ForbiddenAddress(ctx, addr)
}

func (message *Message) ActiveAddress(ctx context.Context, addr address.Address) (address.Address, error) {
	return message.Internal.ActiveAddress(ctx, addr)
}

func (message *Message) SetSelectMsgNum(ctx context.Context, addr address.Address, num uint64) (address.Address, error) {
	return message.Internal.SetSelectMsgNum(ctx, addr, num)
}

func (message *Message) ClearUnFillMessage(ctx context.Context, addr address.Address) (int, error) {
	return message.Internal.ClearUnFillMessage(ctx, addr)
}

func (message *Message) SetFeeParams(ctx context.Context, addr address.Address, gasOverEstimation float64, maxFee, maxFeeCap string) (address.Address, error) {
	return message.Internal.SetFeeParams(ctx, addr, gasOverEstimation, maxFee, maxFeeCap)
}

/////// shared params ///////

func (message *Message) GetSharedParams(ctx context.Context) (*types.SharedParams, error) {
	return message.Internal.GetSharedParams(ctx)
}

func (message *Message) SetSharedParams(ctx context.Context, params *types.SharedParams) (struct{}, error) {
	return message.Internal.SetSharedParams(ctx, params)
}

func (message *Message) RefreshSharedParams(ctx context.Context) (struct{}, error) {
	return message.Internal.RefreshSharedParams(ctx)
}

/////// node info ///////

func (message *Message) SaveNode(ctx context.Context, node *types.Node) (struct{}, error) {
	return message.Internal.SaveNode(ctx, node)
}

func (message *Message) GetNode(ctx context.Context, name string) (*types.Node, error) {
	return message.Internal.GetNode(ctx, name)
}

func (message *Message) HasNode(ctx context.Context, name string) (bool, error) {
	return message.Internal.HasNode(ctx, name)
}

func (message *Message) ListNode(ctx context.Context) ([]*types.Node, error) {
	return message.Internal.ListNode(ctx)
}

func (message *Message) DeleteNode(ctx context.Context, name string) (struct{}, error) {
	return message.Internal.DeleteNode(ctx, name)
}

/////// log ///////

func (message *Message) SetLogLevel(ctx context.Context, level string) error {
	return message.Internal.SetLogLevel(ctx, level)
}

/////// send ///////

func (message *Message) Send(ctx context.Context, params types.SendParams) (string, error) {
	return message.Internal.Send(ctx, params)
}
