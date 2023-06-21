package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/filecoin-project/go-jsonrpc"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/utils"
	"github.com/ipfs-force-community/sophon-messager/filestore"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"

	"github.com/ipfs-force-community/sophon-messager/config"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
)

const (
	OldRepoPath = "~/.venus-messager"
	DefRepoPath = "~/.sophon-messager"
)

func getAPI(ctx *cli.Context) (messager.IMessager, jsonrpc.ClientCloser, error) {
	repo, err := getRepo(ctx)
	if err != nil {
		return nil, func() {}, err
	}
	token, err := repo.GetToken()
	if err != nil {
		return nil, func() {}, err
	}

	cfg := repo.Config()

	return messager.DialIMessagerRPC(ctx.Context, cfg.API.Address, string(token), nil)
}

func getNodeAPI(ctx *cli.Context) (v1.FullNode, jsonrpc.ClientCloser, error) {
	cfg, err := getConfig(ctx)
	if err != nil {
		return nil, func() {}, err
	}
	return v1.DialFullNodeRPC(ctx.Context, cfg.Node.Url, cfg.Node.Token, nil)
}

func NewNodeAPI(ctx context.Context, addr, token string) (v1.FullNode, jsonrpc.ClientCloser, error) {
	return v1.DialFullNodeRPC(ctx, addr, token, nil)
}

func getConfig(ctx *cli.Context) (*config.Config, error) {
	repo, err := getRepo(ctx)
	if err != nil {
		return nil, err
	}

	return repo.Config(), nil
}

func LoadBuiltinActors(ctx context.Context, nodeAPI v1.FullNode) error {
	if err := utils.LoadBuiltinActors(ctx, nodeAPI); err != nil {
		return err
	}
	utils.ReloadMethodsMap()

	return nil
}

func getRepo(ctx *cli.Context) (filestore.FSRepo, error) {
	repoPath, err := GetRepoPath(ctx)
	if err != nil {
		return nil, err
	}
	repo, err := filestore.NewFSRepo(repoPath)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func GetRepoPath(cctx *cli.Context) (string, error) {
	repoPath, err := homedir.Expand(cctx.String("repo"))
	if err != nil {
		return "", err
	}
	has, err := hasFSRepo(repoPath)
	if err != nil {
		return "", err
	}
	if !has {
		// check old repo path
		rPath, err := homedir.Expand(OldRepoPath)
		if err != nil {
			return "", err
		}
		has, err = hasFSRepo(rPath)
		if err != nil {
			return "", err
		}
		if has {
			return rPath, nil
		}
	}
	return repoPath, nil
}

func hasFSRepo(repoPath string) (bool, error) {
	fi, err := os.Stat(repoPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	if !fi.IsDir() {
		return false, fmt.Errorf("%s is not a folder", repoPath)
	}

	return true, nil
}
