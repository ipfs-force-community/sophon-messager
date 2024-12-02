package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	shared "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/ipfs-force-community/sophon-messager/filestore"
	"github.com/ipfs-force-community/sophon-messager/models/mysql"
	"github.com/ipfs-force-community/sophon-messager/models/repo"
	"github.com/ipfs-force-community/sophon-messager/models/sqlite"
	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

var log = logging.Logger("db")

func SetDataBase(fsRepo filestore.FSRepo) (repo.Repo, error) {
	switch fsRepo.Config().DB.Type {
	case "sqlite":
		return sqlite.OpenSqlite(fsRepo)
	case "mysql":
		return mysql.OpenMysql(&fsRepo.Config().DB.MySql)
	default:
		return nil, fmt.Errorf("unexpected db type %s (want 'sqlite' or 'mysql')", fsRepo.Config().DB.Type)
	}
}

func AutoMigrate(repo repo.Repo) error {
	if err := repo.AutoMigrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return MigrateAddress(repo)
}

func Options() fx.Option {
	return fx.Options(
		fx.Provide(SetDataBase),
		fx.Invoke(AutoMigrate),
		// repo
		fx.Provide(repo.NewINodeRepo),
		fx.Provide(repo.NewINodeProvider),
	)
}

func MigrateAddress(r repo.Repo) error {
	list, err := r.AddressRepo().ListAddress(context.Background())
	if err != nil {
		return err
	}

	return r.Transaction(func(txRepo repo.TxRepo) error {
		for _, addrInfo := range list {
			fAddr := addrInfo.Addr.String()
			_, err := txRepo.AddressRepo().GetOneRecord(context.Background(), fAddr)
			if err == nil {
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			tAddr := "t" + fAddr[1:]

			log.Infof("migrate address %s to %s", tAddr, fAddr)
			now := time.Now()
			newAddrInfo := &types.Address{
				ID:        shared.NewUUID(),
				Addr:      addrInfo.Addr,
				Nonce:     addrInfo.Nonce,
				SelMsgNum: addrInfo.SelMsgNum,
				State:     addrInfo.State,
				IsDeleted: repo.NotDeleted,
				FeeSpec:   addrInfo.FeeSpec,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := txRepo.AddressRepo().SaveAddress(context.Background(), newAddrInfo); err != nil {
				return err
			}
			log.Infof("migrate address %s to %s success", tAddr, fAddr)
			if err := txRepo.AddressRepo().DelAddress(context.Background(), tAddr); err != nil {
				return err
			}
			log.Infof("delete address %s success", tAddr)
		}
		return nil
	})
}
