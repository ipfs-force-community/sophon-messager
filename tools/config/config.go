package config

import (
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"
)

type Config struct {
	Venus        VenusConfig
	Messager     MessagerConfig
	BatchReplace BatchReplaceConfig
}

type MessagerConfig struct {
	ConnectConfig
}

type VenusConfig struct {
	ConnectConfig
}

type ConnectConfig struct {
	URL   string
	Token string
}

type BatchReplaceConfig struct {
	BlockTime string
	From      Address
	Selects   []Select
}

type Select struct {
	ActorCode CID
	Methods   []int
}

func DefaultConfig() *Config {
	return &Config{
		Messager: MessagerConfig{
			ConnectConfig: ConnectConfig{
				URL:   "/ip4/127.0.0.1/tcp/39812",
				Token: "",
			},
		},
		Venus: VenusConfig{
			ConnectConfig: ConnectConfig{
				URL:   "/ip4/127.0.0.1/tcp/3453",
				Token: "",
			},
		},
		BatchReplace: BatchReplaceConfig{
			BlockTime: "5m",
			From:      UndefAddress,
			Selects: []Select{
				{
					ActorCode: UndefCID,
					Methods:   []int{5},
				},
			},
		},
	}
}

type CID cid.Cid

var UndefCID = CID(cid.Undef)

func (c *CID) UnmarshalTOML(in interface{}) error {
	var data string
	switch d := in.(type) {
	case string:
		data = d
	case []byte:
		data = string(d)
	default:
		return fmt.Errorf("unexpected type %T", in)
	}
	if len(data) == 0 {
		*c = UndefCID
		return nil
	}
	res, err := cid.Decode(data)
	if err != nil {
		return err
	}
	*c = CID(res)

	return nil
}

func (c CID) MarshalTOML() ([]byte, error) {
	if c == UndefCID {
		return []byte("\"\""), nil
	}
	data := cid.Cid(c).String()

	return []byte(fmt.Sprintf("\"%s\"", data)), nil
}

func (c *CID) Cid() cid.Cid {
	if c == nil {
		return cid.Undef
	}
	return cid.Cid(*c)
}

func (c CID) String() string {
	return cid.Cid(c).String()
}

type Address address.Address

var UndefAddress = Address(address.Undef)

func (a *Address) UnmarshalTOML(in interface{}) error {
	var data string
	switch d := in.(type) {
	case string:
		data = d
	case []byte:
		data = string(d)
	default:
		return fmt.Errorf("unexpected type %T", in)
	}
	if len(data) == 0 {
		*a = UndefAddress
		return nil
	}
	addr, err := address.NewFromString(data)
	if err != nil {
		return err
	}
	*a = Address(addr)

	return nil
}

func (a Address) MarshalTOML() ([]byte, error) {
	if a == UndefAddress {
		return []byte("\"\""), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", address.Address(a).String())), nil
}

func (a *Address) Empty() bool {
	if a == nil {
		return true
	}
	return *a == UndefAddress
}

func (a *Address) Address() address.Address {
	if a == nil {
		return address.Undef
	}
	return address.Address(*a)
}

func (a Address) String() string {
	return address.Address(a).String()
}
