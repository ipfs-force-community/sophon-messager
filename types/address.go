package types

import (
	"fmt"
	"time"

	"github.com/filecoin-project/go-address"
)

type AddressState int

const (
	Alive AddressState = iota
	Notfound
	Removed
	Forbiden
)

//         ------> Forbiden
//         |           |
//  |-- Alive <--------
//  |                  |
//   -> Notfound ---> Remove
//

type Address struct {
	ID   UUID            `json:"id"`
	Addr address.Address `json:"addr"`
	//max for current, use nonce and +1
	Nonce    uint64       `json:"nonce"`
	Weight   int64        `json:"weight"`
	WalletID UUID         `json:"walletID"`
	State    AddressState `json:"state"`
	//number of address selection messages
	SelectMsgNum uint64 `json:"SelectMsgNum"`

	IsDeleted int       `json:"isDeleted"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `json:"createAt"`  // 创建时间
	UpdatedAt time.Time `json:"updateAt"`  // 更新时间
}

func AddrStateToString(state AddressState) string {
	switch state {
	case Alive:
		return "Alive"
	case Notfound:
		return "Notfound"
	case Removed:
		return "Removed"
	case Forbiden:
		return "Forbiden"
	default:
		return fmt.Sprintf("unknow state %d", state)
	}
}
