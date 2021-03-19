package main

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs-force-community/venus-messager/api/client"
	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/types"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	cfg, err := config.ReadConfig("/Users/lijunlong/Desktop/workload/venus-messager/messager.toml")
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
	addr1, _ := address.NewFromString("t3v3wx6tbwlvzev7hxhbpdlfwvwq5mbdhfrgmy2i2ztfaqhwjwc6zkxo6to4x2ms2acicd3x57fabxhpszzwqq")
	addr2, _ := address.NewFromString("t3ru4e5hrvhsjjvyxyzzxzmsoahrdmobsfz6ohmd7ftswxyf7dxvhnmkq63cu5ozdy4wnrcqxx4gkwa427grga")

	addrs := []address.Address{addr1, addr2}
	tm := time.NewTicker(time.Second * 1)
	/*for i:=0 ;i< 100;i++ {
		fmt.Println(rand.Intn(2))
	}*/
	//return
	for {
		select {
		case <-tm.C:
			for i := 0; i < 50; i++ {
				from := addrs[rand.Intn(2)]
				to := addrs[rand.Intn(2)]
				fmt.Println(from)
				uid, err := client.PushMessageWithId(context.Background(),
					types.NewUUID(),
					&venustypes.UnsignedMessage{
						Version: 0,
						To:      from,
						From:    to,
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
		}
	}

	/*msg, err := client.WaitMessage(context.Background(), uid, 5)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("wait for message ", msg.SignedCid)
	fmt.Println("code:", msg.Receipt.ExitCode)
	fmt.Println("gas_used:", msg.Receipt.GasUsed)
	fmt.Println("return_value:", msg.Receipt.ReturnValue)
	fmt.Println("Height:", msg.Height)
	fmt.Println("Tipset:", msg.TipSetKey.String())*/
}
