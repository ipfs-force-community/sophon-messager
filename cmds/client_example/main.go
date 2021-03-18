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

	timer := time.NewTicker(time.Second * 3)
	defer timer.Stop()

	sendMsg := func() {
		addr, _ := address.NewFromString("t3wgwvaidoz27bigs6dcqoutcyf6tqhinzrphbprahrmc7xz3lhilmllkzytmi6sihw6wxmtnaf6eh3knvppzq")
		uid, err := client.PushMessageWithId(context.Background(),
			types.NewUUID(),
			&venustypes.UnsignedMessage{
				Version: 0,
				To:      addr,
				From:    addr,
				Nonce:   1,
				Value:   abi.NewTokenAmount(100),
				Method:  0,
			},
			&types.MsgMeta{
				ExpireEpoch:       1000000,
				GasOverEstimation: 1.25,
				MaxFee:            big.NewInt(10000000000000000),
				MaxFeeCap:         big.NewInt(10000000000000000),
			})
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Println("send message " + uid.String())

	}

	for {
		select {
		case <-timer.C:
			for i := 0; i < 5; i++ {
				sendMsg()
			}
		}
	}

	//msg, err := client.WaitMessage(context.Background(), uid, 5)
	//if err != nil {
	//	log.Fatal(err)
	//	return
	//}
	//
	//fmt.Println("wait for message ", msg.SignedCid)
	//fmt.Println("code:", msg.Receipt.ExitCode)
	//fmt.Println("gas_used:", msg.Receipt.GasUsed)
	//fmt.Println("return_value:", msg.Receipt.ReturnValue)
	//fmt.Println("Height:", msg.Height)
	//fmt.Println("Tipset:", msg.TipSetKey.String())
}
