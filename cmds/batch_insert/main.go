package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/models"
	"github.com/filecoin-project/venus-messager/models/mysql"
	venustypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	"github.com/ipfs/go-cid"
)

var params = `
{
	"SectorNumber": 22000,
	"Proof": "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"
}
`

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	var dns string
	var batch int
	flag.StringVar(&dns, "dns", "root:password@(127.0.0.1:3306)/venus_messager?parseTime=true&loc=Local", "")
	flag.IntVar(&batch, "batch", 300, "")

	flag.Parse()

	cfg := &config.MySqlConfig{
		ConnectionString: dns,
		MaxOpenConn:      100,
		MaxIdleConn:      100,
		ConnMaxLifeTime:  time.Hour,
	}
	r, err := mysql.OpenMysql(cfg)
	checkErr(err)
	checkErr(r.AutoMigrate())

	msgCID, err := cid.Decode("bafy2bzacebaajxztamw3odd5nlfvv62le67nw53afgkmuarierdbrken7ebcy")
	checkErr(err)

	cids := make([]cid.Cid, 0, 10)
	for i := 0; i < 10; i++ {
		cids = append(cids, msgCID)
	}
	key := venustypes.NewTipSetKey(cids...)

	for {
		msgs := make([]*types.Message, 0, batch)
		for i := 0; i < batch; i++ {
			msg := models.NewMessage()
			msgcid := msg.Cid()
			msg.Params = []byte(params)
			msg.Confidence = 10
			msg.CreatedAt = time.Now()
			msg.UpdatedAt = time.Now()
			msg.Height = rand.Int63n(2000000)
			msg.WalletName = "admin"
			msg.FromUser = "admin"
			msg.UnsignedCid = &msgcid
			msg.SignedCid = &msgcid
			msg.Signature = &crypto.Signature{
				Type: crypto.SigTypeBLS,
				Data: []byte(msgCID.String()),
			}
			msg.Receipt = &venustypes.MessageReceipt{
				ExitCode: -1,
				Return:   []byte{},
				GasUsed:  1000000000,
			}
			msg.TipSetKey = key

			msgs = append(msgs, msg)
		}
		start := time.Now()
		checkErr(r.MessageRepo().BatchSaveMessage(msgs))
		fmt.Printf("batch %d spent %v'ms'", batch, time.Since(start).Milliseconds())
	}
}
