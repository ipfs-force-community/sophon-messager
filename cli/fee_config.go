package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/filecoin-project/go-state-types/big"
	venusTypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
	"gorm.io/gorm"

	"github.com/filecoin-project/venus-messager/types"
)

var methodString = `
Global or Wallet:        -1
Send:                     0
Constructor:              1
ControlAddresses:         2
ChangeWorkerAddress:      3
ChangePeerID:             4
SubmitWindowedPoSt:       5
PreCommitSector:          6
ProveCommitSector:        7
ExtendSectorExpiration:   8
TerminateSectors:         9
DeclareFaults:            10
DeclareFaultsRecovered:   11
OnDeferredCronEvent:      12
CheckSectorProven:        13
ApplyRewards:             14
ReportConsensusFault:     15
WithdrawBalance:          16
ConfirmSectorProofsValid: 17
ChangeMultiaddrs:         18
CompactPartitions:        19
CompactSectorNumbers:     20
ConfirmUpdateWorkerKey:   21
RepayDebt:                22
ChangeOwnerAddress:       23
DisputeWindowedPoSt:      24
`

var FeeConfigCmds = &cli.Command{
	Name:  "fee-config",
	Usage: "fee config commands",
	Subcommands: []*cli.Command{
		setFeeConfigCmd,
		searchFeeConfigCmd,
		listFeeConfigCmd,
		delFeeConfigCmd,
	},
}

var walletNameFlag = &cli.StringFlag{
	Name:  "wallet-name",
	Usage: "wallet name",
}
var methodFlag = &cli.Int64Flag{
	Name: "method",
	Usage: `
The value corresponding to message method, -1 is an extra and represents a global or wallet level configuration 
eg.` + methodString,
}

var setFeeConfigCmd = &cli.Command{
	Name:  "set",
	Usage: "set fee config, eg. ./venus-messager fee-config set --gas-overestimation 1.25 --max-feecap 0 --max-fee 70000000000000000 --global",
	Flags: []cli.Flag{
		walletNameFlag,
		methodFlag,
		&cli.Float64Flag{
			Name:  "gas-overestimation",
			Usage: "Estimate the coefficient of gas",
		},
		&cli.StringFlag{
			Name:  "max-feecap",
			Usage: "Max feecap for a message (burn and pay to miner, attoFIL/GasUnit)",
		},
		&cli.StringFlag{
			Name:  "max-fee",
			Usage: "Spend up to X attoFIL for message",
		},
		&cli.BoolFlag{
			Name:    "global",
			Usage:   "Whether global configuration",
			Aliases: []string{"g"},
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		var feeConfig types.FeeConfig
		walletName := ctx.String("wallet-name")
		method := ctx.Int64("method")

		if ctx.Bool("global") {
			oldFeeConfig, err := client.GetGlobalFeeConfig(ctx.Context)
			if err == nil {
				feeConfig = *oldFeeConfig
			} else if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
				feeConfig.ID = types.DefGlobalFeeCfgID
				feeConfig.CreatedAt = time.Now()
				feeConfig.IsDeleted = -1
			} else {
				return err
			}
		} else {
			wallet, err := client.GetWalletByName(ctx.Context, walletName)
			if err != nil {
				return xerrors.Errorf("found wallet %v", err)
			}
			oldFeeConfig, err := client.GetFeeConfig(ctx.Context, wallet.ID, method)
			if err == nil {
				feeConfig = *oldFeeConfig
			} else if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
				feeConfig.ID = types.NewUUID()
				feeConfig.Method = method
				feeConfig.WalletID = wallet.ID
				feeConfig.IsDeleted = -1
			} else {
				return err
			}
		}
		if ctx.IsSet("gas-overestimation") {
			feeConfig.GasOverEstimation = ctx.Float64("gas-overestimation")
		}
		if ctx.IsSet("max-fee") {
			feeConfig.MaxFee, err = venusTypes.BigFromString(ctx.String("max-fee"))
			if err != nil {
				return xerrors.Errorf("parsing max-spend: %w", err)
			}
		}
		if ctx.IsSet("max-feecap") {
			feeConfig.MaxFeeCap, err = venusTypes.BigFromString(ctx.String("max-feecap"))
			if err != nil {
				return xerrors.Errorf("parsing max-feecap: %w", err)
			}
		}
		uuid, err := client.SaveFeeConfig(ctx.Context, &feeConfig)
		fmt.Println(uuid)

		return err
	},
}

var searchFeeConfigCmd = &cli.Command{
	Name:  "search",
	Usage: "search fee config",
	Flags: []cli.Flag{
		walletNameFlag,
		methodFlag,
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		walletName := ctx.String("wallet-name")
		method := ctx.Int64("method")
		wallet, err := client.GetWalletByName(ctx.Context, walletName)
		if err != nil {
			return err
		}
		fcs, err := client.GetFeeConfig(ctx.Context, wallet.ID, method)
		if err != nil {
			return err
		}

		bytes, err := json.MarshalIndent(transformFeeConfig(fcs, walletName), " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var listFeeConfigCmd = &cli.Command{
	Name:  "list",
	Usage: "list fee config",
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		fcs, err := client.ListFeeConfig(ctx.Context)
		if err != nil {
			return err
		}

		feeCfgs := make([]feeConfig, 0, len(fcs))
		for _, fc := range fcs {
			walletName := ""
			if fc.ID == types.DefGlobalFeeCfgID {
				walletName = `""` + "(global fee config)"
			} else {
				wallet, err := client.GetWalletByID(ctx.Context, fc.WalletID)
				if err == nil {
					walletName = wallet.Name
				} else if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
					walletName = `""` + "(wallet maybe deleted)"
				} else {
					return xerrors.Errorf("found wallet(%v) %v", fc.WalletID, err)
				}
			}
			feeCfgs = append(feeCfgs, transformFeeConfig(fc, walletName))
		}

		bytes, err := json.MarshalIndent(feeCfgs, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var delFeeConfigCmd = &cli.Command{
	Name:  "del",
	Usage: "delete one fee config",
	Flags: []cli.Flag{
		walletNameFlag,
		methodFlag,
		&cli.BoolFlag{
			Name:    "del-global",
			Usage:   "delete global fee config",
			Aliases: []string{"dg"},
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		var walletID types.UUID
		method := ctx.Int64("method")
		walletName := ctx.String("wallet-name")
		if ctx.Bool("del-global") {
			walletID = types.UUID{}
			method = -1
		} else {
			wallet, err := client.GetWalletByName(ctx.Context, walletName)
			if err != nil {
				return xerrors.Errorf("found wallet %v", err)
			}
			walletID = wallet.ID
		}
		_, err = client.DeleteFeeConfig(ctx.Context, walletID, method)
		if err != nil {
			return err
		}

		fmt.Println("success")
		return nil
	},
}

type feeConfig struct {
	ID                types.UUID
	WalletName        string
	Method            int64
	GasOverEstimation float64
	MaxFee            big.Int
	MaxFeeCap         big.Int

	IsDeleted int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func transformFeeConfig(fc *types.FeeConfig, walletName string) feeConfig {
	if fc == nil {
		return feeConfig{}
	}

	return feeConfig{
		ID:                fc.ID,
		WalletName:        walletName,
		Method:            fc.Method,
		GasOverEstimation: fc.GasOverEstimation,
		MaxFee:            fc.MaxFee,
		MaxFeeCap:         fc.MaxFeeCap,
		IsDeleted:         fc.IsDeleted,
		CreatedAt:         fc.CreatedAt,
		UpdatedAt:         fc.UpdatedAt,
	}
}
