package cli

import (
	"encoding/json"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs-force-community/venus-messager/types"

	venustypes "github.com/filecoin-project/venus/pkg/types"

	"github.com/urfave/cli/v2"
)

var MsgCmds = &cli.Command{
	Name:  "msg",
	Usage: "msg commands",
	Subcommands: []*cli.Command{
		pushCmd,
		getCmd,
		listCmd,
	},
}

var pushCmd = &cli.Command{
	Name:  "push",
	Usage: "list local msg test",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "uuid",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		to, err := address.NewFromString("f01000")
		if err != nil {
			return err
		}
		from, err := address.NewFromString("f01002")
		if err != nil {
			return err
		}
		msg := &types.Message{
			ID:          types.NewUUID(),
			UnsignedCid: nil,
			SignedCid:   nil,
			UnsignedMessage: venustypes.UnsignedMessage{
				Version:    0,
				To:         to,
				From:       from,
				Nonce:      0,
				Value:      abi.TokenAmount{},
				GasLimit:   0,
				GasFeeCap:  abi.TokenAmount{},
				GasPremium: abi.TokenAmount{},
				Method:     0,
				Params:     nil,
			},
			Signature: nil,
			Height:    0,
			Receipt:   nil,
			Meta:      nil,
			State:     0,
		}
		id, err := client.PushMessage(ctx.Context, msg)
		if err != nil {
			return err
		}
		fmt.Println(id)
		return nil
	},
}

var getCmd = &cli.Command{
	Name:  "get",
	Usage: "get local msg test",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "uuid",
		},
		&cli.StringFlag{
			Name:  "output-type",
			Usage: "output type support json and table",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		uid := ctx.String("uuid")
		msg, err := client.GetMessage(ctx.Context, uid)
		if err != nil {
			return err
		}
		bytes, err := json.MarshalIndent(msg, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}

var listCmd = &cli.Command{
	Name:  "list",
	Usage: "list local msg test",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "output-type",
			Usage: "output type support json and table",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, closer, err := getAPI(ctx)
		if err != nil {
			return err
		}
		defer closer()

		msg, err := client.ListMessage(ctx.Context)
		if err != nil {
			return err
		}
		bytes, err := json.MarshalIndent(msg, " ", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytes))
		return nil
	},
}
