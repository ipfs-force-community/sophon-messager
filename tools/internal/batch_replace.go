package internal

import (
	"fmt"
	"sort"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	cli2 "github.com/filecoin-project/venus-messager/cli"
	"github.com/filecoin-project/venus-messager/tools/config"
	"github.com/filecoin-project/venus-messager/utils"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/urfave/cli/v2"
)

var BatchReplaceCmd = &cli.Command{
	Name:  "batch-replace",
	Usage: "batch replace messages",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "gas-feecap",
			Usage: "gas feecap for new message (burn and pay to miner, attoFIL/GasUnit)",
		},
		&cli.StringFlag{
			Name:  "gas-premium",
			Usage: "gas premium for new message (pay to miner, attoFIL/GasUnit)",
		},
		cli2.GasOverPremiumFlag,
		&cli.Int64Flag{
			Name:  "gas-limit",
			Usage: "gas limit for new message (GasUnit)",
		},
		&cli.BoolFlag{
			Name:  "auto",
			Usage: "automatically reprice the specified message",
		},
		&cli.StringFlag{
			Name:  "max-fee",
			Usage: "Spend up to X attoFIL for this message (applicable for auto mode)",
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context
		cfg := new(config.Config)
		err := utils.ReadConfig(cctx.String("config"), cfg)
		if err != nil {
			return fmt.Errorf("read config failed: %v", err)
		}

		messagerAPI, closer, err := cli2.NewMessagerAPI(ctx, cfg.Messager.URL, cfg.Messager.Token)
		if err != nil {
			return err
		}
		defer closer()

		nodeAPI, closer2, err := cli2.NewNodeAPI(ctx, cfg.Venus.URL, cfg.Venus.Token)
		if err != nil {
			return err
		}
		defer closer2()

		blockTime, err := time.ParseDuration(cfg.BatchReplace.BlockTime)
		if err != nil {
			return err
		}

		// check
		if cctx.IsSet("gas-premium") && cctx.IsSet(cli2.GasOverPremiumFlag.Name) {
			return fmt.Errorf("gas-premium and gas-over-premium flag only need one")
		}
		if len(cfg.BatchReplace.Selects) == 0 {
			return fmt.Errorf("selects are empty, please check config")
		}

		params, err := cli2.ParseFlagToReplaceMessaeParams(cctx)
		if err != nil {
			return err
		}

		addrs := make(map[address.Address]struct{})
		if !cfg.BatchReplace.From.Empty() {
			addrs[cfg.BatchReplace.From.Address()] = struct{}{}
		} else {
			addrList, err := messagerAPI.ListAddress(ctx)
			if err != nil {
				return err
			}
			for _, addrInfo := range addrList {
				addrs[addrInfo.Addr] = struct{}{}
			}
		}

		pendingMsgs := make(map[address.Address][]*messager.Message, len(addrs))
		for addr := range addrs {
			actor, err := nodeAPI.StateGetActor(ctx, addr, types.EmptyTSK)
			if err != nil {
				return fmt.Errorf("get actor %s failed: %v", addr, err)
			}

			blockedMsgs, err := messagerAPI.ListBlockedMessage(ctx, addr, blockTime)
			if err != nil {
				return err
			}

			sort.Slice(blockedMsgs, func(i, j int) bool {
				return blockedMsgs[i].Nonce < blockedMsgs[j].Nonce
			})

			tmsgs, err := selectMsg(cfg.BatchReplace.Selects, addr, actor, blockedMsgs)
			if err != nil {
				fmt.Printf("address %s selset message failed: %v", addr, err)
				continue
			}

			pendingMsgs[addr] = tmsgs
			fmt.Printf("address %s has %d message need replace\n", addr, len(tmsgs))
		}

		if len(pendingMsgs) == 0 {
			return nil
		}

		for addr, msgs := range pendingMsgs {
			fmt.Printf("\nstart replace message for %s \n", addr)
			for _, msg := range msgs {
				params.ID = msg.ID
				oldCid := msg.Cid()
				newCid, err := messagerAPI.ReplaceMessage(cctx.Context, params)
				if err != nil {
					fmt.Printf("replace msg %s %d failed %v\n", msg.ID, msg.Nonce, err)
					continue
				}
				fmt.Printf("replace message %s success, old cid: %s, new cid: %s\n", msg.ID, oldCid, newCid)
			}
			fmt.Printf("end replace message for %s\n", addr)
		}

		return nil
	},
}

func selectMsg(sels []config.Select, addr address.Address, actor *types.Actor, msgs []*messager.Message) ([]*messager.Message, error) {
	var newSels []config.Select
	for _, sel := range sels {
		if sel.ActorCode.Cid() == actor.Code {
			newSels = append(newSels, sel)
		}
	}
	if len(newSels) == 0 {
		fmt.Printf("address %s(%s) not match actor code\n", addr, actor.Code)
		return nil, nil
	}
	if len(newSels) > 1 {
		return nil, fmt.Errorf("repeat actor code %v", newSels[0].ActorCode.Cid())
	}

	res := make([]*messager.Message, 0, len(msgs))
	methods := make(map[abi.MethodNum]struct{}, len(newSels[0].Methods))
	for _, m := range newSels[0].Methods {
		methods[abi.MethodNum(m)] = struct{}{}
	}
	for _, msg := range msgs {
		if _, ok := methods[msg.Message.Method]; !ok {
			continue
		}
		if msg.Nonce < actor.Nonce {
			continue
		}
		res = append(res, msg)
	}

	return res, nil
}
