package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"

	actors2 "github.com/filecoin-project/venus/venus-shared/actors"

	"github.com/filecoin-project/venus/venus-shared/utils"

	"github.com/filecoin-project/go-state-types/abi"

	"github.com/filecoin-project/go-state-types/actors"
	types2 "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/venus-messager/cli/tablewriter"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/urfave/cli/v2"
)

var ActorCfgCmds = &cli.Command{
	Name:  "actor",
	Usage: "actor config",
	Subcommands: []*cli.Command{
		listActorCfgCmd,
		getActorCfgCmd,
		updateActorCfgCmd,
		addActorCfgCmd,
		listBuiltinActorCmd,
	},
}

var listActorCfgCmd = &cli.Command{
	Name:  "list",
	Usage: "list actor config",
	Flags: []cli.Flag{
		outputTypeFlag,
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		nodeAPI, nodeAPICloser, err := getNodeAPI(ctx)
		if err != nil {
			return err
		}
		defer nodeAPICloser()

		if err := LoadBuiltinActors(ctx.Context, nodeAPI); err != nil {
			return err
		}

		actorCfgs, err := client.ListActorCfg(ctx.Context)
		if err != nil {
			return err
		}

		if ctx.String(outputTypeFlag.Name) == "table" {
			return outputActorCfgWithTable(actorCfgs)
		}

		bytes, err := json.MarshalIndent(actorCfgs, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var updateActorCfgCmd = &cli.Command{
	Name:      "update-fee-params",
	Usage:     "update fee params of actor config",
	ArgsUsage: "<uid>",
	Flags: []cli.Flag{
		gasOverEstimationFlag,
		gasFeeCapFlag,
		maxFeeFlag,
		basefeeFlag,
		GasOverPremiumFlag,
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		nodeAPI, nodeAPICloser, err := getNodeAPI(ctx)
		if err != nil {
			return err
		}
		defer nodeAPICloser()

		if err := LoadBuiltinActors(ctx.Context, nodeAPI); err != nil {
			return err
		}

		if ctx.NArg() != 1 {
			return errors.New("must specify one uid argument")
		}
		uuidStr := ctx.Args().Get(0)
		id, err := types2.ParseUUID(uuidStr)
		if err != nil {
			return err
		}

		hasUpdate := false
		changeGasSpecParams := &types.ChangeGasSpecParams{}
		if ctx.IsSet(gasOverEstimationFlag.Name) {
			hasUpdate = true
			gasOverEstimation := ctx.Float64(gasOverEstimationFlag.Name)
			changeGasSpecParams.GasOverEstimation = &gasOverEstimation
		}

		if ctx.IsSet(GasOverPremiumFlag.Name) {
			hasUpdate = true
			gasOverPremium := ctx.Float64(GasOverPremiumFlag.Name)
			changeGasSpecParams.GasOverPremium = &gasOverPremium
		}

		if ctx.IsSet(gasFeeCapFlag.Name) {
			hasUpdate = true
			gasFeeCap, err := types2.BigFromString(ctx.String(gasFeeCapFlag.Name))
			if err != nil {
				return fmt.Errorf("parsing feecap failed %v", err)
			}
			changeGasSpecParams.GasFeeCap = gasFeeCap
		}

		if ctx.IsSet(maxFeeFlag.Name) {
			hasUpdate = true
			maxfee, err := types2.BigFromString(ctx.String(maxFeeFlag.Name))
			if err != nil {
				return fmt.Errorf("parsing maxfee failed %v", err)
			}
			changeGasSpecParams.MaxFee = maxfee
		}

		if ctx.IsSet(basefeeFlag.Name) {
			hasUpdate = true
			basefee, err := types2.BigFromString(ctx.String(basefeeFlag.Name))
			if err != nil {
				return fmt.Errorf("parsing maxfee failed %v", err)
			}
			changeGasSpecParams.BaseFee = basefee
		}
		if !hasUpdate {
			return errors.New("no new params to update")
		}
		return client.UpdateActorCfg(ctx.Context, id, changeGasSpecParams)
	},
}

var addActorCfgCmd = &cli.Command{
	Name:      "add",
	Usage:     "add new actor config",
	ArgsUsage: "<code> <method>",
	Flags: []cli.Flag{
		gasOverEstimationFlag,
		gasFeeCapFlag,
		maxFeeFlag,
		basefeeFlag,
		GasOverPremiumFlag,
		&cli.IntFlag{
			Name: "version",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		nodeAPI, nodeAPICloser, err := getNodeAPI(ctx)
		if err != nil {
			return err
		}
		defer nodeAPICloser()

		if err := LoadBuiltinActors(ctx.Context, nodeAPI); err != nil {
			return err
		}

		if ctx.NArg() != 2 {
			return errors.New("must specify code and method argument")
		}
		codeCid, err := cid.Decode(ctx.Args().Get(0))
		if err != nil {
			return err
		}

		method, err := strconv.ParseUint(ctx.Args().Get(1), 10, 64)
		if err != nil {
			return err
		}
		newActorCfg := &types.ActorCfg{
			ID: types2.NewUUID(),
			MethodType: types.MethodType{
				Code:   codeCid,
				Method: abi.MethodNum(method),
			},
			FeeSpec: types.FeeSpec{},
		}

		if ctx.IsSet("version") {
			newActorCfg.ActorVersion = actors.Version(ctx.Int("version"))
		} else {
			//try resolve builtin
			methodMeta, found := utils.MethodsMap[newActorCfg.Code][newActorCfg.Method]
			if !found {
				return fmt.Errorf("actor not found:%v method(%d) not exist", newActorCfg.Code, newActorCfg.Method)
			}
			newActorCfg.ActorVersion = methodMeta.Version
		}

		hasArgument := false
		if ctx.IsSet(gasOverEstimationFlag.Name) {
			hasArgument = true
			gasOverEstimation := ctx.Float64(gasOverEstimationFlag.Name)
			newActorCfg.GasOverEstimation = gasOverEstimation
		}

		if ctx.IsSet(GasOverPremiumFlag.Name) {
			hasArgument = true
			gasOverPremium := ctx.Float64(GasOverPremiumFlag.Name)
			newActorCfg.GasOverPremium = gasOverPremium
		}

		if ctx.IsSet(gasFeeCapFlag.Name) {
			hasArgument = true
			gasFeeCap, err := types2.BigFromString(ctx.String(gasFeeCapFlag.Name))
			if err != nil {
				return fmt.Errorf("parsing feecap failed %v", err)
			}
			newActorCfg.GasFeeCap = gasFeeCap
		}

		if ctx.IsSet(maxFeeFlag.Name) {
			hasArgument = true
			maxfee, err := types2.BigFromString(ctx.String(maxFeeFlag.Name))
			if err != nil {
				return fmt.Errorf("parsing maxfee failed %v", err)
			}
			newActorCfg.MaxFee = maxfee
		}

		if ctx.IsSet(basefeeFlag.Name) {
			hasArgument = true
			basefee, err := types2.BigFromString(ctx.String(basefeeFlag.Name))
			if err != nil {
				return fmt.Errorf("parsing maxfee failed %v", err)
			}
			newActorCfg.BaseFee = basefee
		}
		if !hasArgument {
			return errors.New("no argument to save")
		}
		return client.SaveActorCfg(ctx.Context, newActorCfg)
	},
}

var getActorCfgCmd = &cli.Command{
	Name:      "get",
	Usage:     "get actor config",
	ArgsUsage: "<uid>",
	Flags: []cli.Flag{
		outputTypeFlag,
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		nodeAPI, nodeAPICloser, err := getNodeAPI(ctx)
		if err != nil {
			return err
		}
		defer nodeAPICloser()

		if err := LoadBuiltinActors(ctx.Context, nodeAPI); err != nil {
			return err
		}

		if ctx.NArg() != 1 {
			return errors.New("must specific one uid argument")
		}
		uuidStr := ctx.Args().Get(0)
		id, err := types2.ParseUUID(uuidStr)
		if err != nil {
			return err
		}

		actorCfg, err := client.GetActorCfgByID(ctx.Context, id)
		if err != nil {
			return err
		}

		if ctx.String(outputTypeFlag.Name) == "table" {
			return outputActorCfgWithTable([]*types.ActorCfg{actorCfg})
		}

		bytes, err := json.MarshalIndent(actorCfg, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var listBuiltinActorCmd = &cli.Command{
	Name:  "list-builtin-actors",
	Usage: "list builtin actors",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name: "version",
		},
	},
	Action: func(ctx *cli.Context) error {
		type MethodMeta struct {
			Name    string
			Code    cid.Cid
			Method  abi.MethodNum
			Version actors.Version
		}

		nodeAPI, nodeAPICloser, err := getNodeAPI(ctx)
		if err != nil {
			return err
		}
		defer nodeAPICloser()

		if err := LoadBuiltinActors(ctx.Context, nodeAPI); err != nil {
			return err
		}

		methodsByActorsVersion := make(map[actors.Version]map[cid.Cid][]MethodMeta)
		var maxVersion actors.Version
		for codeCid, methodMap := range utils.MethodsMap {
			for methodNumber, meta := range methodMap {
				_, ok := methodsByActorsVersion[meta.Version]
				if !ok {
					methodsByActorsVersion[meta.Version] = make(map[cid.Cid][]MethodMeta)
				}
				methods := append(methodsByActorsVersion[meta.Version][codeCid], MethodMeta{
					Name:    meta.Name,
					Code:    codeCid,
					Method:  methodNumber,
					Version: meta.Version,
				})
				sort.Slice(methods, func(i, j int) bool {
					return methods[i].Method < methods[j].Method
				})
				methodsByActorsVersion[meta.Version][codeCid] = methods
				if meta.Version > maxVersion {
					maxVersion = meta.Version
				}
			}
		}

		var version actors.Version
		if ctx.IsSet("version") {
			version = actors.Version(ctx.Int("version"))
		} else {
			//get max version
			version = maxVersion
		}

		var actorCfgTw = tablewriter.New(
			tablewriter.Col("Category"),
			tablewriter.Col("Version"),
			tablewriter.Col("Code"),
			tablewriter.Col("Method"),
			tablewriter.Col("MethodName"),
		)
		printBuiltinTypes := methodsByActorsVersion[version]
		for codeCid, methods := range printBuiltinTypes {
			name, _, has := actors2.GetActorMetaByCode(codeCid)
			if !has {
				return fmt.Errorf("not found builtin actor %s", codeCid)
			}
			actorCfgTw.Write(map[string]interface{}{
				"Category": name,
			})
			for _, method := range methods {
				row := map[string]interface{}{
					"Version":    method.Version,
					"Code":       codeCid,
					"Method":     method.Method,
					"MethodName": method.Name,
				}
				actorCfgTw.Write(row)
			}
		}
		return actorCfgTw.Flush(os.Stdout)
	},
}
var actorCfgTw = tablewriter.New(
	tablewriter.Col("ID"),
	tablewriter.Col("NVersion"),
	tablewriter.Col("Code"),
	tablewriter.Col("Method"),
	tablewriter.Col("MethodName"),
	tablewriter.Col("GasOverEstimation"),
	tablewriter.Col("GasOverPremium"),
	tablewriter.Col("MaxFee"),
	tablewriter.Col("GasFeeCap"),
	tablewriter.Col("BaseFee"),
	tablewriter.Col("CreateAt"),
)

func outputActorCfgWithTable(actorCfg []*types.ActorCfg) error {
	for _, actorCfg := range actorCfg {
		row := map[string]interface{}{
			"ID":                actorCfg.ID,
			"AVersion":          actorCfg.ActorVersion,
			"Code":              actorCfg.Code,
			"Method":            actorCfg.Method,
			"MethodName":        resolveBuiltinMethodName(actorCfg.MethodType),
			"GasOverEstimation": actorCfg.FeeSpec.GasOverEstimation,
			"GasOverPremium":    actorCfg.FeeSpec.GasOverPremium,
			"MaxFee":            actorCfg.FeeSpec.MaxFee,
			"GasFeeCap":         actorCfg.FeeSpec.GasFeeCap,
			"BaseFee":           actorCfg.FeeSpec.BaseFee,
			"CreateAt":          actorCfg.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		actorCfgTw.Write(row)
	}

	buf := new(bytes.Buffer)
	if err := actorCfgTw.Flush(buf); err != nil {
		return err
	}
	fmt.Println(buf)
	return nil
}
