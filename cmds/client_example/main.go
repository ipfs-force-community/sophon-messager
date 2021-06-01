package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	venustypes "github.com/filecoin-project/venus/pkg/types"

	"github.com/filecoin-project/venus-messager/api/client"
	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/types"
	"github.com/filecoin-project/venus-messager/utils"
)

func main() {
	var apiAddress string
	var token string
	var fromStr string
	var toStr string
	var count int
	var value int64

	flag.StringVar(&apiAddress, "api", "", "messager api address")
	flag.StringVar(&token, "token", "", "messager token")
	flag.StringVar(&fromStr, "from", "", "from which address is the message sent")
	flag.StringVar(&toStr, "to", "", "to whom is the message sent")
	flag.IntVar(&count, "count", 50, "number of messages sent per second")
	flag.Int64Var(&value, "value", 0, "")

	flag.Parse()

	from, err := address.NewFromString(fromStr)
	if err != nil {
		panic(err)
	}
	var to address.Address
	if len(toStr) == 0 {
		to = from
	} else {
		to, err = address.NewFromString(toStr)
		if err != nil {
			panic(err)
		}
	}
	if count < 0 {
		count = 1
	}

	cfg, err := config.ReadConfig("./messager.toml")
	if err != nil {
		log.Fatal(err)
		return
	}
	if len(apiAddress) == 0 {
		apiAddress = cfg.API.Address
	}

	fmt.Println("api address: ", apiAddress)
	fmt.Println("token      : ", token)
	fmt.Println("from       : ", from.String())
	fmt.Println("to         : ", to.String())
	fmt.Println("count      : ", count)
	fmt.Println("value      : ", value)

	addr, err := utils.DialArgs(apiAddress)
	if err != nil {
		log.Fatal(err)
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	client, closer, err := client.NewMessageRPC(context.Background(), addr, header)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer closer()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for i := 0; i < count; i++ {
			msgMate := &types.MsgMeta{
				GasOverEstimation: 1.25,
				MaxFee:            big.NewInt(10000000000000000),
				MaxFeeCap:         big.NewInt(10000000000000000),
			}
			uid, err := client.PushMessageWithId(context.Background(),
				types.NewUUID().String(),
				&venustypes.UnsignedMessage{
					Version: 0,
					To:      to,
					From:    from,
					Nonce:   1,
					Value:   abi.NewTokenAmount(value),
					Method:  0,
				},
				msgMate,
			)
			if err != nil {
				log.Fatal(err)
				return
			}

			fmt.Println("send message " + uid)
		}
	}
}
