package mtypes

import (
	"database/sql/driver"
	"errors"

	"github.com/ipfs/go-cid"
)

type DBCid cid.Cid

var UndefDBCid = DBCid{}

func NewDBCid(id cid.Cid) DBCid {
	return DBCid(id)
}

func (c *DBCid) Scan(value interface{}) error {
	var val string
	switch value := value.(type) {
	case []byte:
		val = string(value)
	case string:
		val = value
	default:
		return errors.New("address should be a `[]byte` or `string`")
	}

	if len(val) == 0 {
		*c = UndefDBCid
		return nil
	}
	cid, err := cid.Decode(val)
	if err != nil {
		return err
	}
	*c = DBCid(cid)

	return nil
}

func (c DBCid) Value() (driver.Value, error) {
	return c.String(), nil
}

func (c DBCid) String() string {
	if c == UndefDBCid {
		return ""
	}
	return cid.Cid(c).String()
}

func (c DBCid) Cid() cid.Cid {
	return cid.Cid(c)
}

func (c DBCid) cidPtr() *cid.Cid {
	if c == UndefDBCid {
		return nil
	}
	cid := cid.Cid(c)
	return &cid
}
