package main

import (
	"context"
	"github.com/ipfs-force-community/venus-messager/api/client"
	"github.com/ipfs-force-community/venus-messager/types"
	"log"
	"net/http"
)

func main() {
	header := http.Header{}
	header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoibGkiLCJwZXJtIjoiYWRtaW4iLCJleHQiOiJleUpCYkd4dmR5STZXeUp5WldGa0lpd2lkM0pwZEdVaUxDSnphV2R1SWl3aVlXUnRhVzRpWFgwIn0.BBkE5F-Z0NgUcvCsj7CYFdcef92NAvdUuWbUSHpew0E")
	client, closer, err := client.NewMessageRPC(context.Background(), "http://127.0.0.1:39812/rpc/v0", header)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer closer()

	_, err = client.SetSharedParams(context.Background(), &types.SharedParams{
		ID:                 1,
		ExpireEpoch:        1,
		GasOverEstimation:  1,
		MaxFee:             1,
		MaxFeeCap:          1,
		SelMsgNum:          1,
		ScanInterval:       1,
		MaxEstFailNumOfMsg: 1,
	})

	if err != nil {
		log.Fatal(err)
		return
	}
}
