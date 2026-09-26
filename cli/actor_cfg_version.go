package cli

import (
	"fmt"
	"os"
	"sort"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/actors"
	actors2 "github.com/filecoin-project/venus/venus-shared/actors"
	"github.com/filecoin-project/venus/venus-shared/api/messager"
	types2 "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/filecoin-project/venus/venus-shared/utils"
	"github.com/ipfs-force-community/sophon-messager/cli/tablewriter"
	"github.com/ipfs/go-cid"
	"github.com/urfave/cli/v2"
)

// Actor config rows are keyed by the on-chain actor code CID, which the chain
// hands to `getActorCfg` in service/message_selector.go. An upgrade that ships
// new builtin actor code CIDs therefore leaves every stored row keyed to a
// superseded CID: the lookups miss and the actor level gas parameters silently
// stop applying. These commands report that gap and move the rows to the code
// CIDs of the new actor version.

type actorCfgVersionGap struct {
	ActorName     string
	Method        abi.MethodNum
	MethodName    string
	Current       *types.ActorCfg
	TargetCode    cid.Cid
	TargetVersion actors.Version
	HasTargetRow  bool
}

type actorCfgVersionStatus struct {
	TargetVersion actors.Version

	// Superseded holds configs keyed to a code CID that the target actor
	// version replaced, together with the code CID that replaces it.
	Superseded []actorCfgVersionGap
	Current    int

	// Foreign holds configs whose code CID is not a builtin actor code of the
	// queried network, NoTarget those of a builtin actor that the target actor
	// version does not ship.
	Foreign  []*types.ActorCfg
	NoTarget []*types.ActorCfg
}

// pending returns the superseded configs that have no counterpart at the target
// actor version yet.
func (s *actorCfgVersionStatus) pending() []actorCfgVersionGap {
	pending := make([]actorCfgVersionGap, 0, len(s.Superseded))
	for _, gap := range s.Superseded {
		if !gap.HasTargetRow {
			pending = append(pending, gap)
		}
	}

	return pending
}

type actorCfgMethodKey struct {
	code   cid.Cid
	method abi.MethodNum
}

// builtinActorCodesByVersion indexes the embedded builtin actor metadata of a
// network as actor name -> code CID, per actor version.
func builtinActorCodesByVersion(network string) map[actors.Version]map[string]cid.Cid {
	byVersion := make(map[actors.Version]map[string]cid.Cid)
	for _, meta := range actors2.EmbeddedBuiltinActorsMetadata {
		if meta.Network == network {
			byVersion[meta.Version] = meta.Actors
		}
	}

	return byVersion
}

// newestActorVersion returns the newest actor version the binary knows for the
// network, which is the version a released binary is upgraded to.
func newestActorVersion(network string) (actors.Version, error) {
	newest, found := actors.Version(0), false
	for version := range builtinActorCodesByVersion(network) {
		if !found || version > newest {
			newest, found = version, true
		}
	}
	if !found {
		return 0, fmt.Errorf("no builtin actor metadata for network %q", network)
	}

	return newest, nil
}

// analyzeActorCfgVersion classifies actor configs against the code CIDs of the
// target actor version.
func analyzeActorCfgVersion(cfgs []*types.ActorCfg, network string, target actors.Version) (*actorCfgVersionStatus, error) {
	byVersion := builtinActorCodesByVersion(network)
	targetCodes, ok := byVersion[target]
	if !ok {
		return nil, fmt.Errorf("no builtin actor metadata for network %q actor version %d", network, target)
	}

	nameByCode := make(map[cid.Cid]string)
	for _, codes := range byVersion {
		for name, code := range codes {
			nameByCode[code] = name
		}
	}

	configured := make(map[actorCfgMethodKey]struct{}, len(cfgs))
	for _, cfg := range cfgs {
		configured[actorCfgMethodKey{code: cfg.Code, method: cfg.Method}] = struct{}{}
	}

	status := &actorCfgVersionStatus{TargetVersion: target}
	for _, cfg := range cfgs {
		name, builtin := nameByCode[cfg.Code]
		if !builtin {
			status.Foreign = append(status.Foreign, cfg)
			continue
		}
		targetCode, ok := targetCodes[name]
		if !ok {
			status.NoTarget = append(status.NoTarget, cfg)
			continue
		}
		if targetCode == cfg.Code {
			status.Current++
			continue
		}

		_, hasTargetRow := configured[actorCfgMethodKey{code: targetCode, method: cfg.Method}]
		status.Superseded = append(status.Superseded, actorCfgVersionGap{
			ActorName:     name,
			Method:        cfg.Method,
			MethodName:    builtinMethodName(cfg.Code, cfg.Method),
			Current:       cfg,
			TargetCode:    targetCode,
			TargetVersion: target,
			HasTargetRow:  hasTargetRow,
		})
	}
	sort.Slice(status.Superseded, func(i, j int) bool {
		if status.Superseded[i].ActorName != status.Superseded[j].ActorName {
			return status.Superseded[i].ActorName < status.Superseded[j].ActorName
		}

		return status.Superseded[i].Method < status.Superseded[j].Method
	})

	return status, nil
}

func builtinMethodName(code cid.Cid, method abi.MethodNum) string {
	meta, ok := utils.MethodsMap[code][method]
	if !ok {
		return ""
	}

	return meta.Name
}

func resolveTargetActorVersion(ctx *cli.Context, network string) (actors.Version, error) {
	if version := ctx.Int("version"); version != 0 {
		return actors.Version(version), nil
	}

	return newestActorVersion(network)
}

func actorCfgVersionStatusFor(ctx *cli.Context, client messager.IMessager) (*actorCfgVersionStatus, error) {
	target, err := resolveTargetActorVersion(ctx, actors2.NetworkBundle)
	if err != nil {
		return nil, err
	}

	cfgs, err := client.ListActorCfg(ctx.Context)
	if err != nil {
		return nil, err
	}

	return analyzeActorCfgVersion(cfgs, actors2.NetworkBundle, target)
}

func outputActorCfgVersionStatus(network string, status *actorCfgVersionStatus) error {
	fmt.Printf("network %s, actor version %d\n", network, status.TargetVersion)
	fmt.Printf("actor configs: %d current, %d superseded, %d of another network, %d without actor version %d code\n",
		status.Current, len(status.Superseded), len(status.Foreign), len(status.NoTarget), status.TargetVersion)
	if len(status.Superseded) == 0 {
		return nil
	}

	actorCfgVersionTw := tablewriter.New(
		tablewriter.Col("Actor"),
		tablewriter.Col("Method"),
		tablewriter.Col("MethodName"),
		tablewriter.Col("ConfiguredCode"),
		tablewriter.Col("TargetCode"),
		tablewriter.Col("TargetRow"),
	)
	for _, gap := range status.Superseded {
		targetRow := "missing"
		if gap.HasTargetRow {
			targetRow = "present"
		}
		actorCfgVersionTw.Write(map[string]interface{}{
			"Actor":          gap.ActorName,
			"Method":         gap.Method,
			"MethodName":     gap.MethodName,
			"ConfiguredCode": gap.Current.Code.String(),
			"TargetCode":     gap.TargetCode.String(),
			"TargetRow":      targetRow,
		})
	}

	return actorCfgVersionTw.Flush(os.Stdout)
}

var verifyActorCfgVersionCmd = &cli.Command{
	Name:  "verify-version",
	Usage: "check that actor configs are keyed to the builtin actor code CIDs of the current actor version",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "version",
			Usage: "actor version to check against, defaults to the newest the binary knows for the network",
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

		status, err := actorCfgVersionStatusFor(ctx, client)
		if err != nil {
			return err
		}

		if err := outputActorCfgVersionStatus(actors2.NetworkBundle, status); err != nil {
			return err
		}

		if pending := status.pending(); len(pending) > 0 {
			return fmt.Errorf("%d actor config(s) are keyed to superseded code CIDs and no longer apply; run `messager actor migrate-version`", len(pending))
		}

		return nil
	},
}

var migrateActorCfgVersionCmd = &cli.Command{
	Name:  "migrate-version",
	Usage: "add an actor config for every superseded code CID at the current actor version, keeping the configured gas parameters",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "version",
			Usage: "actor version to migrate to, defaults to the newest the binary knows for the network",
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

		status, err := actorCfgVersionStatusFor(ctx, client)
		if err != nil {
			return err
		}

		if err := outputActorCfgVersionStatus(actors2.NetworkBundle, status); err != nil {
			return err
		}

		pending := status.pending()
		if len(pending) == 0 {
			fmt.Println("nothing to migrate")

			return nil
		}

		for _, gap := range pending {
			cfg := &types.ActorCfg{
				ID:           types2.NewUUID(),
				ActorVersion: gap.TargetVersion,
				MethodType:   types.MethodType{Code: gap.TargetCode, Method: gap.Method},
				FeeSpec:      gap.Current.FeeSpec,
			}
			if err := client.SaveActorCfg(ctx.Context, cfg); err != nil {
				return fmt.Errorf("save actor config for %s method %d: %w", gap.ActorName, gap.Method, err)
			}
			fmt.Printf("migrated %s method %d (%s) %s -> %s\n",
				gap.ActorName, gap.Method, gap.MethodName, gap.Current.Code, gap.TargetCode)
		}
		fmt.Printf("migrated %d actor config(s) to actor version %d\n", len(pending), status.TargetVersion)

		return nil
	},
}
