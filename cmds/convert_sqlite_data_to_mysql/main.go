package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/filecoin-project/go-state-types/abi"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

// go run cmds/convert_sqlite_data_to_mysql/main.go '{"addr":"127.0.0.1:3306","user":"root","pass":"Root1234","name":"messager","maxOpenConn":500,"maxIdleConn":500,"connMaxLifeTime":600,"debug":false}' message.db
func main() {
	mysqlCfg := &config.DbConfig{
		Type: "mysql",
		MySql: config.MySqlConfig{
			Addr:            "127.0.0.1:3306",
			User:            "root",
			Pass:            "Root1234",
			Name:            "messager",
			MaxOpenConn:     500,
			MaxIdleConn:     500,
			ConnMaxLifeTime: 600,
			Debug:           false,
		},
	}
	sqlitCfg := &config.DbConfig{
		Type: "sqlite",
		Sqlite: config.SqliteConfig{
			Path:  "message.db",
			Debug: false,
		},
	}

	args := os.Args
	fmt.Println("args: ", len(args), args)
	if len(args) == 3 {
		fmt.Println("args: ", args)
		var mysqlConfig config.MySqlConfig
		if err := json.Unmarshal([]byte(args[1]), &mysqlConfig); err != nil {
			log.Fatal(err)
		}
		mysqlCfg.MySql = mysqlConfig

		sqlitCfg.Sqlite.Path = args[2]
	}
	cfgByte, err := json.Marshal(mysqlCfg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("sqlite config: ", sqlitCfg)
	fmt.Println("mysql config: ", string(cfgByte))

	sqliteRepo, err := models.SetDataBase(sqlitCfg)
	if err != nil {
		log.Fatal(err)
	}
	if err = sqliteRepo.AutoMigrate(); err != nil {
		log.Fatal(err)
	}
	mysqlRepo, err := models.SetDataBase(mysqlCfg)
	if err != nil {
		log.Fatal(err)
	}
	if err = mysqlRepo.AutoMigrate(); err != nil {
		log.Fatal(err)
	}
	msgs, wallets, addrs, walletAddrs, nodes, sharedParams := loadDataFromSqlite(sqliteRepo)

	saveDataToMysql(mysqlRepo, msgs, wallets, addrs, walletAddrs, nodes, sharedParams)
}

func saveDataToMysql(mysqlRepo repo.Repo, msgs []*types.Message, wallets []*types.Wallet, addrs []*types.Address, walletAddrs []*types.WalletAddress, nodes []*types.Node, shareParams *types.SharedParams) {
	for _, msg := range msgs {
		if err := mysqlRepo.MessageRepo().CreateMessage(msg); err != nil {
			log.Fatal(err)
		}
		// When you use the create command, the default value is used instead of the zero value.
		// So msg.Receipt.ExitCode need save again
		if msg.State == 3 && msg.Receipt.ExitCode == 0 {
			if err := mysqlRepo.MessageRepo().UpdateMessageInfoByCid(msg.UnsignedCid.String(), msg.Receipt, abi.ChainEpoch(msg.Height), msg.State, msg.TipSetKey); err != nil {
				log.Fatal(err)
			}
		}
	}
	for _, wallet := range wallets {
		if err := mysqlRepo.WalletRepo().SaveWallet(wallet); err != nil {
			log.Fatal(err)
		}
	}
	ctx := context.TODO()
	for _, addr := range addrs {
		if err := mysqlRepo.AddressRepo().SaveAddress(ctx, addr); err != nil {
			log.Fatal(err)
		}
	}
	for _, wa := range walletAddrs {
		if err := mysqlRepo.WalletAddressRepo().SaveWalletAddress(wa); err != nil {
			log.Fatal(err)
		}
	}
	for _, node := range nodes {
		if err := mysqlRepo.NodeRepo().SaveNode(node); err != nil {
			log.Fatal(err)
		}
	}
	if _, err := mysqlRepo.SharedParamsRepo().SetSharedParams(ctx, shareParams); err != nil {
		log.Fatal(err)
	}
}

func loadDataFromSqlite(repo repo.Repo) ([]*types.Message, []*types.Wallet, []*types.Address, []*types.WalletAddress, []*types.Node, *types.SharedParams) {
	ctx := context.TODO()
	msgs, err := repo.MessageRepo().ListMessage()
	if err != nil {
		log.Fatal(err)
	}
	wallets, err := repo.WalletRepo().ListWallet()
	if err != nil {
		log.Fatal(err)
	}
	addrs, err := repo.AddressRepo().ListAddress(ctx)
	if err != nil {
		log.Fatal(err)
	}
	walletAddrs, err := repo.WalletAddressRepo().ListWalletAddress()
	if err != nil {
		log.Fatal(err)
	}
	nodes, err := repo.NodeRepo().ListNode()
	if err != nil {
		log.Fatal(err)
	}
	sharedParams, err := repo.SharedParamsRepo().GetSharedParams(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return msgs, wallets, addrs, walletAddrs, nodes, sharedParams
}
