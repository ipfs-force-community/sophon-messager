package service

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/big"
	chain2 "github.com/filecoin-project/venus/app/submodule/chain"
	"github.com/filecoin-project/venus/pkg/chain"
	"github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"net/http"
	"net/url"
)

type NodeClient struct {
	ChainNotify            func(context.Context) (<-chan []*chain.HeadChange, error)
	ChainHead              func(context.Context) (*types.TipSet, error)
	ChainGetTipSet         func(context.Context, types.TipSetKey) (*types.TipSet, error)
	ChainGetBlock          func(context.Context, cid.Cid) (*types.BlockHeader, error)
	ChainGetBlockMessages  func(context.Context, cid.Cid) (*chain2.BlockMessages, error)
	ChainGetParentMessages func(ctx context.Context, bcid cid.Cid) ([]chain2.Message, error)
	ChainGetParentReceipts func(context.Context, cid.Cid) ([]*types.MessageReceipt, error)
	StateAccountKey        func(ctx context.Context, addr address.Address, tsk types.TipSetKey) (address.Address, error)
	StateSearchMsg         func(context.Context, cid.Cid) (*chain.MsgLookup, error)
	StateGetActor          func(context.Context, address.Address, types.TipSetKey) (*types.Actor, error)

	GasEstimateMessageGas func(context.Context, *types.UnsignedMessage, *types.MessageSendSpec, types.TipSetKey) (*types.UnsignedMessage, error)
	GasEstimateFeeCap     func(context.Context, *types.UnsignedMessage, int64, types.TipSetKey) (big.Int, error)
	GasEstimateGasPremium func(context.Context, uint64, address.Address, int64, types.TipSetKey) (big.Int, error)
	GasEstimateGasLimit   func(ctx context.Context, msgIn *types.UnsignedMessage, tsk types.TipSetKey) (int64, error)
	//	BatchGasEstimateMessageGas func(ctx context.Context, estimateMessages []*types.EstimateMessage, tsk types.TipSetKey) ([]*types.UnsignedMessage, error)

	MpoolPush      func(context.Context, *types.SignedMessage) (cid.Cid, error)
	MpoolBatchPush func(context.Context, []*types.SignedMessage) ([]cid.Cid, error)
}

func NewNodeClient(ctx context.Context, cfg *config.NodeConfig) (*NodeClient, jsonrpc.ClientCloser, error) {
	headers := http.Header{}
	if len(cfg.Token) != 0 {
		headers.Add("Authorization", "Bearer "+string(cfg.Token))
	}
	addr, err := DialArgs(cfg.Url)
	if err != nil {
		return nil, nil, err
	}
	var res NodeClient
	closer, err := jsonrpc.NewMergeClient(ctx, addr, "Filecoin", []interface{}{&res}, headers)
	return &res, closer, err
}

func DialArgs(addr string) (string, error) {
	ma, err := multiaddr.NewMultiaddr(addr)
	if err == nil {
		_, addr, err := manet.DialArgs(ma)
		if err != nil {
			return "", err
		}

		return "ws://" + addr + "/rpc/v0", nil
	}

	_, err = url.Parse(addr)
	if err != nil {
		return "", err
	}
	return addr + "/rpc/v0", nil
}
