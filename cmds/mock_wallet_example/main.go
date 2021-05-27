package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-messager/api/client"
	"github.com/filecoin-project/venus-wallet/core"
	"github.com/filecoin-project/venus/pkg/crypto"
	"github.com/google/uuid"
	gatewayTypes "github.com/ipfs-force-community/venus-gateway/types"
)

// go run cmds/mock_wallet_example/main.go --account testminer --private-key 7b22707269766174654b6579223a226b45564e4d662b48533242593469774c7535374f37675055625449776859504e4a364264717532556e47453d222c2274797065223a22626c73227d
func main() {
	var privateKeyString string
	var addr address.Address
	var account string
	var accounts []string

	flag.StringVar(&privateKeyString, "private-key", "", "private key")
	flag.StringVar(&account, "account", "", "account")

	flag.Parse()

	ki, err := convertToKeyInfo(privateKeyString)
	if err != nil {
		panic(err)
	}
	addr, err = ki.Address()
	if err != nil {
		panic(err)
	}
	accounts = append(accounts, strings.Split(account, ",")...)
	fmt.Printf("current address: %s, account: %s \n", addr.String(), accounts)

	ws := &WalletService{
		ki:      ki,
		addr:    addr,
		account: account,
	}

	header := http.Header{}
	gatewayCli, closer, err := client.NewMessageRPC(context.Background(), "ws://127.0.0.1:39812/rpc/v0", header)
	if err != nil {
		panic(err)
	}
	defer closer()

	reqEvent, err := gatewayCli.ListenWalletEvent(context.Background(), accounts)
	if err != nil {
		panic(err)
	}

	toResponseEvent := func(id uuid.UUID, payload []byte, err error) *gatewayTypes.ResponseEvent {
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		return &gatewayTypes.ResponseEvent{
			Id:      id,
			Payload: payload,
			Error:   errStr,
		}
	}

	go func() {
		for e := range reqEvent {
			fmt.Println("receive request: ", *e)
			switch e.Method {
			case "InitConnect":
				//req := gatewayTypes.ConnectedCompleted{}
				//err := json.Unmarshal(event.Payload, &req)
				//if err != nil {
				//gatewayCli.ResponseEvent(context.Background(), toResponseEvent(e.Id, nil, err))
				//}
			case "WalletList":
				list, _ := ws.WalletList(context.Background())
				b, err := json.Marshal(list)
				if err := gatewayCli.ResponseEvent(context.Background(), toResponseEvent(e.Id, b, err)); err != nil {
					fmt.Println("WalletList:ResponseEvent failed", err)
				}
				fmt.Println("call walletlist ", list)
			case "WalletSign":
				var wsr gatewayTypes.WalletSignRequest
				var resp *gatewayTypes.ResponseEvent
				err := json.Unmarshal(e.Payload, &wsr)
				if err != nil {
					resp = toResponseEvent(e.Id, []byte{}, err)
				} else {
					signature, err := ws.WalletSign(context.Background(), "", wsr.Signer, wsr.ToSign, wsr.Meta)
					if err != nil {
						resp = toResponseEvent(e.Id, []byte{}, err)
					} else {
						b, err := json.Marshal(signature)
						resp = toResponseEvent(e.Id, b, err)
					}
				}
				if err := gatewayCli.ResponseEvent(context.Background(), resp); err != nil {
					fmt.Println("WalletSign:ResponseEvent failed", err)
				}
			default:
				fmt.Printf("invalid method %v\n", e.Method)
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

func convertToKeyInfo(privateKeyStr string) (*crypto.KeyInfo, error) {
	var ki crypto.KeyInfo
	b, err := hex.DecodeString(privateKeyStr)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &ki)
	if err != nil {
		return nil, err
	}

	return &ki, err
}

type WalletService struct {
	ki      *crypto.KeyInfo
	addr    address.Address
	account string
}

func (ws *WalletService) WalletHas(ctx context.Context, supportAccount string, addr address.Address) (bool, error) {
	if ws.account != supportAccount || ws.addr != addr {
		return false, nil
	}
	return true, nil
}

func (ws *WalletService) WalletList(ctx context.Context) ([]address.Address, error) {
	return []address.Address{ws.addr}, nil
}

func (ws *WalletService) WalletSign(ctx context.Context, account string, addr address.Address, toSign []byte, meta core.MsgMeta) (*crypto.Signature, error) {
	sig, err := crypto.Sign(toSign, ws.ki.PrivateKey, ws.ki.SigType)
	return sig, err
}
