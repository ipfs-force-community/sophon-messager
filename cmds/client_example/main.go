package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus/pkg/constants"
	venustypes "github.com/filecoin-project/venus/venus-shared/types"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

func main() {
	var apiAddress string
	var token string
	var fromStr string
	var toStr string
	var count int
	var value string

	flag.StringVar(&apiAddress, "api", "/ip4/127.0.0.1/tcp/39812", "messager api address")
	flag.StringVar(&token, "token", "", "messager token")
	flag.StringVar(&fromStr, "from", "", "from which address is the message sent")
	flag.StringVar(&toStr, "to", "", "to whom is the message sent")
	flag.IntVar(&count, "count", 1, "number of messages sent per second")
	flag.StringVar(&value, "value", "1", "")

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

	client, closer, err := messager.DialIMessagerRPC(context.Background(), apiAddress, token, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer closer()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for i := 0; i < count; i++ {
			msgMate := &types.SendSpec{
				GasOverEstimation: 1.25,
				MaxFee:            big.NewInt(10000000000000000),
			}
			uid, err := client.PushMessageWithId(context.Background(),
				venustypes.NewUUID().String(),
				&venustypes.Message{
					Version: 0,
					To:      to,
					From:    from,
					Nonce:   0,
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

			go func() {
				msg, err := client.WaitMessage(context.Background(), uid, constants.MessageConfidence)
				if err != nil {
					return
				}
				fmt.Printf("wait message success: %s, cid: %s\n", uid, msg.SignedCid)
			}()
		}
	}
}
