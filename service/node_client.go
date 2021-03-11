package service

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	chain2 "github.com/filecoin-project/venus/app/submodule/chain"
	"github.com/filecoin-project/venus/app/submodule/network"
	"github.com/filecoin-project/venus/pkg/chain"
	"github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	"github.com/ipfs-force-community/venus-messager/config"
)

type NodeClient struct {
	ChainNotify            func(context.Context) (<-chan []*chain.HeadChange, error)
	BlockTime              func(context.Context) time.Duration
	ChainHead              func(context.Context) (*types.TipSet, error)
	ChainList              func(context.Context, types.TipSetKey, int) ([]types.TipSetKey, error)
	ChainSetHead           func(context.Context, types.TipSetKey) error
	ChainGetTipSet         func(context.Context, types.TipSetKey) (*types.TipSet, error)
	ChainGetTipSetByHeight func(context.Context, abi.ChainEpoch, types.TipSetKey) (*types.TipSet, error)
	ChainGetBlock          func(context.Context, cid.Cid) (*types.BlockHeader, error)
	ChainGetMessage        func(context.Context, cid.Cid) (*types.UnsignedMessage, error)
	ChainGetBlockMessages  func(context.Context, cid.Cid) (*chain2.BlockMessages, error)
	ChainGetReceipts       func(context.Context, cid.Cid) ([]types.MessageReceipt, error)
	ChainGetParentMessages func(ctx context.Context, bcid cid.Cid) ([]chain2.Message, error)
	ChainGetParentReceipts func(context.Context, cid.Cid) ([]*types.MessageReceipt, error)
	GetFullBlock           func(context.Context, cid.Cid) (*types.FullBlock, error)
	GetActor               func(context.Context, address.Address) (*types.Actor, error)
	GetEntry               func(context.Context, abi.ChainEpoch, uint64) (*types.BeaconEntry, error)
	MessageWait            func(context.Context, cid.Cid, abi.ChainEpoch, abi.ChainEpoch) (*chain.ChainMessage, error)
	ResolveToKeyAddr       func(context.Context, address.Address, *types.TipSet) (address.Address, error)
	StateNetworkName       func(context.Context) (chain2.NetworkName, error)
	StateSearchMsg         func(context.Context, cid.Cid) (*chain.MsgLookup, error)
	StateNetworkVersion    func(context.Context, types.TipSetKey) (network.Version, error)
	StateGetActor          func(context.Context, address.Address, types.TipSetKey) (*types.Actor, error)
	StateSearchMsgLimited  func(context.Context, cid.Cid, abi.ChainEpoch) (*chain.MsgLookup, error)
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
