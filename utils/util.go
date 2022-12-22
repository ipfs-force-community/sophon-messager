package utils

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/go-address"
	"github.com/pelletier/go-toml"
)

func StringToTipsetKey(str string) (types.TipSetKey, error) {
	str = strings.TrimLeft(str, "{ ")
	str = strings.TrimRight(str, " }")
	var cids []cid.Cid
	for _, s := range strings.Split(str, " ") {
		c, err := cid.Decode(s)
		if err != nil {
			return types.TipSetKey{}, err
		}
		cids = append(cids, c)
	}

	return types.NewTipSetKey(cids...), nil
}

// WriteJson original data will be cleared
func WriteJson(filePath string, obj interface{}) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0o666)
	if err != nil {
		return err
	}
	defer file.Close() // nolint

	b, err := json.MarshalIndent(obj, " ", "\t")
	if err != nil {
		return err
	}
	_, err = file.Write(b)

	return err
}

func ReadConfig(path string, cfg interface{}) error {
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return toml.Unmarshal(configBytes, cfg)
}

func WriteConfig(path string, cfg interface{}) error {
	cfgBytes, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, cfgBytes, 0o666)
}

func MsgsGroupByAddress(msgs []*types.SignedMessage) map[address.Address][]*types.SignedMessage {
	msgMap := make(map[address.Address][]*types.SignedMessage)
	for _, msg := range msgs {
		msgMap[msg.Message.From] = append(msgMap[msg.Message.From], msg)
	}
	return msgMap
}
