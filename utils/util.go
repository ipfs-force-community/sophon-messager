package utils

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"

	venusTypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs/go-cid"
)

func StringToTipsetKey(str string) (venusTypes.TipSetKey, error) {
	str = strings.TrimLeft(str, "{ ")
	str = strings.TrimRight(str, " }")
	var cids []cid.Cid
	for _, s := range strings.Split(str, " ") {
		c, err := cid.Decode(s)
		if err != nil {
			return venusTypes.TipSetKey{}, err
		}
		cids = append(cids, c)
	}

	return venusTypes.NewTipSetKey(cids...), nil
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
	file, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(file)

	return b, err
}

// original data will be cleared
func WriteFile(filePath string, obj interface{}) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := json.MarshalIndent(obj, " ", "\t")
	if err != nil {
		return err
	}
	_, err = file.Write(b)

	return err
}
