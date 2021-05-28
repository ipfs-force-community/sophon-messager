package types

import (
	"fmt"
	"time"

	"github.com/filecoin-project/go-address"
)

type State int

const (
	_ State = iota
	Alive
	Removing
	Removed
	Forbiden // forbiden received message
)

type Address struct {
	ID   UUID            `json:"id"`
	Addr address.Address `json:"addr"`
	//max for current, use nonce and +1
	Nonce  uint64 `json:"nonce"`
	Weight int64  `json:"weight"`
	// number of address selection messages
	SelMsgNum uint64 `json:"selMsgNum"`
	State     State  `json:"state"`

	IsDeleted int       `json:"isDeleted"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `json:"createAt"`  // 创建时间
	UpdatedAt time.Time `json:"updateAt"`  // 更新时间
}

func StateToString(state State) string {
	switch state {
	case Alive:
		return "Alive"
	case Removing:
		return "Removing"
	case Removed:
		return "Removed"
	case Forbiden:
		return "Forbiden"
	default:
		return fmt.Sprintf("unknow state %d", state)
	}
}
