package utils

import (
	"encoding/json"
	"io"
	"net"
	"os"
	"strings"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/gbrlsnchs/jwt/v3"
	"github.com/ipfs/go-cid"
	"github.com/pelletier/go-toml"
	"golang.org/x/xerrors"
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

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

var _ io.ReadCloser = (*CloserReader)(nil)

type CloserReader struct {
	reader io.Reader
}

func NewCloserReader(reader io.Reader) *CloserReader {
	return &CloserReader{reader: reader}
}

func (c *CloserReader) Read(p []byte) (n int, err error) {
	return c.reader.Read(p)
}

func (c *CloserReader) Close() error {
	return nil
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ReadFile(filePath string) ([]byte, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0o666)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(file)

	return b, err
}

// WriteFile original data will be cleared
func WriteFile(filePath string, obj interface{}) error {
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

func GenToken(pl interface{}) (string, error) {
	secret, err := jwtclient.RandSecret()
	if err != nil {
		return "", xerrors.Errorf("rand secret %v", err)
	}
	tk, err := jwt.Sign(pl, jwt.NewHS256(secret))
	if err != nil {
		return core.EmptyString, xerrors.Errorf("gen token failed :%s", err)
	}
	return string(tk), nil
}
