package cli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/ipfs/go-cid"
	"github.com/urfave/cli/v2"
)

var batchReplaceCmd = &cli.Command{
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
		gasOverPremiumFlag,
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
		&cli.StringFlag{
			Name:  "from",
			Usage: "from which address is the message sent",
		},
		&cli.StringFlag{
			Name:  "block-time",
			Usage: "message retention time in messager",
			Value: "5m",
		},
		&cli.StringFlag{
			Name: "methods",
			Usage: `Choose some methods,
methods:
  5:  SubmitWindowedPoSt
  6:  PreCommitSector
  7:  ProveCommitSector
  11: DeclareFaultsRecovered
`,
			Value: "5",
		},
		&cli.StringFlag{
			Name:     "actor-code",
			Usage:    "use to check address",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := cctx.Context

		client, closer, err := getAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		nodeAPI, closer2, err := getNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer2()

		// check
		if cctx.IsSet("gas-premium") && cctx.IsSet(gasOverPremiumFlag.Name) {
			return fmt.Errorf("gas-premium and gas-over-premium flag only need one")
		}

		params, err := parseFlagToReplaceMessaeParams(cctx)
		if err != nil {
			return err
		}

		methods, err := parseMethods(cctx.String("methods"))
		if err != nil {
			return err
		}

		blockTime, err := time.ParseDuration(cctx.String("block-time"))
		if err != nil {
			return err
		}

		addrs := make(map[address.Address]struct{})
		if cctx.IsSet("from") {
			from, err := address.NewFromString(cctx.String("from"))
			if err != nil {
				return err
			}
			addrs[from] = struct{}{}
		} else {
			addrList, err := client.ListAddress(ctx)
			if err != nil {
				return err
			}
			for _, addrInfo := range addrList {
				addrs[addrInfo.Addr] = struct{}{}
			}
		}

		aimActorCode, err := cid.Parse(cctx.String("actor-code"))
		if err != nil {
			return fmt.Errorf("parse actor code failed %v", err)
		}

		pendingMsgs := make(map[address.Address][]*messager.Message, len(addrs))
		for addr := range addrs {
			actor, err := nodeAPI.StateGetActor(ctx, addr, types.EmptyTSK)
			if err != nil {
				return fmt.Errorf("get actor %s failed: %v", addr, err)
			}
			if !aimActorCode.Equals(actor.Code) {
				fmt.Printf("address %s(%s) not match actor code\n", addr, actor.Code)
				continue
			}

			blockedMsgs, err := client.ListBlockedMessage(ctx, addr, blockTime)
			if err != nil {
				return err
			}

			sort.Slice(blockedMsgs, func(i, j int) bool {
				return blockedMsgs[i].Nonce < blockedMsgs[j].Nonce
			})

			tmsgs := make([]*messager.Message, 0, len(blockedMsgs))
			for _, msg := range blockedMsgs {
				if _, ok := methods[msg.Method]; ok && msg.Nonce >= actor.Nonce {
					tmsgs = append(tmsgs, msg)
				}
			}

			pendingMsgs[addr] = tmsgs
			fmt.Printf("address %s has %d message need replace\n", addr, len(tmsgs))
		}

		for addr, msgs := range pendingMsgs {
			fmt.Printf("\nstart replace message for %s \n", addr)
			for _, msg := range msgs {
				params.ID = msg.ID
				oldCid := msg.Cid()
				newCid, err := client.ReplaceMessage(cctx.Context, params)
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

func parseMethods(mstr string) (map[abi.MethodNum]struct{}, error) {
	methods := make(map[abi.MethodNum]struct{})
	for _, m := range strings.Split(mstr, ",") {
		i, err := strconv.Atoi(m)
		if err != nil {
			return nil, err
		}
		methods[abi.MethodNum(i)] = struct{}{}
	}

	return methods, nil
}

func parseFlagToReplaceMessaeParams(cctx *cli.Context) (*messager.ReplacMessageParams, error) {
	params := messager.ReplacMessageParams{
		Auto:           cctx.Bool("auto"),
		GasLimit:       cctx.Int64("gas-limit"),
		GasOverPremium: cctx.Float64(gasOverPremiumFlag.Name),
	}

	if cctx.IsSet("max-fee") {
		maxFee, err := types.ParseFIL(cctx.String("max-fee"))
		if err != nil {
			return nil, fmt.Errorf("parse max fee failed: %v", err)
		}
		params.MaxFee = big.Int(maxFee)
	}
	if cctx.IsSet("gas-premium") {
		gasPremium, err := types.BigFromString(cctx.String("gas-premium"))
		if err != nil {
			return nil, fmt.Errorf("parse gas premium failed: %v", err)
		}
		params.GasPremium = gasPremium
	}
	if cctx.IsSet("gas-feecap") {
		gasFeecap, err := types.BigFromString(cctx.String("gas-feecap"))
		if err != nil {
			return nil, fmt.Errorf("parse gas feecap failed: %v", err)
		}
		params.GasFeecap = gasFeecap
	}

	return &params, nil
}
