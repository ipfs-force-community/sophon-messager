package types

import (
	"time"

	"github.com/filecoin-project/go-address"
)

type Address struct {
	ID   UUID            `json:"id"`
	Addr address.Address `json:"addr"`
	//max for current, use nonce and +1
	Nonce  uint64 `json:"nonce"`
	Weight int64  `json:"weight"`

	IsDeleted int       `json:"isDeleted"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `json:"createAt"`  // 创建时间
	UpdatedAt time.Time `json:"updateAt"`  // 更新时间
}
