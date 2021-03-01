package node

import (
	"context"
	"github.com/filecoin-project/venus/pkg/chain"
	"github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs-force-community/venus-messager/api/controller"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"golang.org/x/xerrors"
)

type NodeEvents struct {
	client        NodeClient
	msgController controller.Message
}

func StartNodeEvents(lc fx.Lifecycle, client NodeClient, r repo.Repo, log *logrus.Logger) *NodeEvents {
	nd := &NodeEvents{
		client: client,
		msgController: controller.Message{
			controller.BaseController{
				Repo:   r,
				Logger: log,
			},
		},
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			notifs, err := nd.client.ChainNotify(ctx)
			if err != nil {
				return err
			}
			noti := <-notifs
			if len(noti) != 1 {
				return xerrors.Errorf("expect hccurrent length 1 but for %d", len(noti))
			}

			if noti[0].Type != chain.HCCurrent {
				return xerrors.Errorf("expect hccurrent event but got %s ", noti[0].Type)
			}
			//todo do some check or repaire for the first connect
			nd.ProcessHCCurrent(ctx, noti[0].Val)
			go func() {
				for notif := range notifs {
					var apply []*types.TipSet
					var revert []*types.TipSet

					for _, change := range notif {
						switch change.Type {
						case chain.HCApply:
							apply = append(apply, change.Val)
						case chain.HCRevert:
							revert = append(revert, change.Val)
						}
					}

					nd.ProcessEvent(ctx, apply, revert)
				}
			}()
			return nil
		},
	})
	return nd
}

//todo do some check or repaire for the first connect
func (nd *NodeEvents) ProcessHCCurrent(ctx context.Context, head *types.TipSet) {

}

//todo update message status when head change
func (nd *NodeEvents) ProcessEvent(ctx context.Context, apply, revert []*types.TipSet) {

}
