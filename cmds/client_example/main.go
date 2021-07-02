package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ipfs-force-community/venus-common-utils/apiinfo"
	"log"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	venustypes "github.com/filecoin-project/venus/pkg/types"

	"github.com/filecoin-project/venus-messager/api/client"
	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/types"
)

func main() {
	var apiAddress string
	var token string
	var fromStr string
	var toStr string
	var count int
	var value string

	flag.StringVar(&apiAddress, "api", "", "messager api address")
	flag.StringVar(&token, "token", "", "messager token")
	flag.StringVar(&fromStr, "from", "", "from which address is the message sent")
	flag.StringVar(&toStr, "to", "", "to whom is the message sent")
	flag.IntVar(&count, "count", 50, "number of messages sent per second")
	flag.StringVar(&value, "value", "0", "")

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

	if len(token) == 0 {
		token = cfg.Node.Token
	}

	val, err := venustypes.ParseFIL(value)
	if err != nil {
		log.Fatalf("failed to parse amount: %v", err)
		return
	}

	fmt.Println("api address: ", apiAddress)
	fmt.Println("token      : ", token)
	fmt.Println("from       : ", from.String())
	fmt.Println("to         : ", to.String())
	fmt.Println("count      : ", count)
	fmt.Println("value      : ", value)

	apiInfo := apiinfo.NewAPIInfo(apiAddress, token)
	addr, err := apiInfo.DialArgs("v0")
	if err != nil {
		log.Fatal(err)
	}
	client, closer, err := client.NewMessageRPC(context.Background(), addr, apiInfo.AuthHeader())
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
					Value:   abi.TokenAmount(val),
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
