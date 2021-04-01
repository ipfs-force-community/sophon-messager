package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs-force-community/venus-messager/api/client"
	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/types"
)

func main() {
	cfg, err := config.ReadConfig("./messager.toml")
	if err != nil {
		log.Fatal(err)
		return
	}

	header := http.Header{}
	client, closer, err := client.NewMessageRPC(context.Background(), "http://"+cfg.API.Address+"/rpc/v0", header)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer closer()

	//loopPushMsgs(client)

	from, _ := address.NewFromString("t3v3wx6tbwlvzev7hxhbpdlfwvwq5mbdhfrgmy2i2ztfaqhwjwc6zkxo6to4x2ms2acicd3x57fabxhpszzwqq")
	to, _ := address.NewFromString("t3ru4e5hrvhsjjvyxyzzxzmsoahrdmobsfz6ohmd7ftswxyf7dxvhnmkq63cu5ozdy4wnrcqxx4gkwa427grga")

	fmt.Println(from)
	msgMate := &types.MsgMeta{
		ExpireEpoch:       abi.ChainEpoch(1000000),
		GasOverEstimation: 1.25,
		MaxFee:            big.NewInt(10000000000000000),
		MaxFeeCap:         big.NewInt(10000000000000000),
	}
	uid, err := client.PushMessageWithId(context.Background(),
		types.NewUUID().String(),
		&venustypes.UnsignedMessage{
			Version: 0,
			To:      from,
			From:    to,
			Nonce:   1,
			Value:   abi.NewTokenAmount(100),
			Method:  0,
		},
		msgMate)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("send message " + uid)

	msg, err := client.WaitMessage(context.Background(), uid, 5)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("wait for message ", msg.SignedCid)
	fmt.Println("code:", msg.Receipt.ExitCode)
	fmt.Println("gas_used:", msg.Receipt.GasUsed)
	fmt.Println("return_value:", msg.Receipt.ReturnValue)
	fmt.Println("Height:", msg.Height)
	fmt.Println("Tipset:", msg.TipSetKey.String())
}

// nolint
func loopPushMsgs(client client.IMessager) {
	from, _ := address.NewFromString("t3qtatmg6tsxolkrbpbb63lexcxgcph4pujowihkayxx23fonnztfspjhviejflu6ssjitqmx3sei5k63ul5la")
	to, _ := address.NewFromString("t3qtatmg6tsxolkrbpbb63lexcxgcph4pujowihkayxx23fonnztfspjhviejflu6ssjitqmx3sei5k63ul5la")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for i := 0; i < 50; i++ {
			msgMate := &types.MsgMeta{
				ExpireEpoch:       abi.ChainEpoch(1000000),
				GasOverEstimation: 1.25,
				MaxFee:            big.NewInt(10000000000000000),
				MaxFeeCap:         big.NewInt(10000000000000000),
			}
			uid, err := client.PushMessageWithId(context.Background(),
				types.NewUUID().String(),
				&venustypes.UnsignedMessage{
					Version: 0,
					To:      from,
					From:    to,
					Nonce:   1,
					Value:   abi.NewTokenAmount(100),
					Method:  0,
				},
				msgMate)
			if err != nil {
				log.Fatal(err)
				return
			}

			fmt.Println("send message " + uid)
		}
	}
}
