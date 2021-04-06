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

//         ------> Forbiden
//         |           |
//  |-- Alive <--------
//  |                  |
//   -> Removing ---> Removed
//

type Wallet struct {
	ID    UUID   `json:"id"` // 主键
	Name  string `json:"name"`
	Url   string `json:"url"`
	Token string `json:"token"`
	State State  `json:"state"`

	IsDeleted int       `json:"isDeleted"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `json:"createAt"`  // 创建时间
	UpdatedAt time.Time `json:"updateAt"`  // 更新时间
}

type WalletAddress struct {
	ID           UUID            `json:"id"` // 主键
	WalletName   string          `json:"walletName"`
	Addr         address.Address `json:"addr"`
	AddressState State           `json:"addressState"`
	// number of address selection messages
	SelMsgNum uint64 `json:"selMsgNum"`

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
