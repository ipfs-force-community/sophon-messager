package types

import "time"

type Wallet struct {
	Id    string `json:"id"` // 主键
	Name  string `json:"name"`
	Url   string `json:"url"`
	Token string `json:"token"`

	IsDeleted int       `json:"isDeleted"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `json:"createAt"`  // 创建时间
	UpdatedAt time.Time `json:"updateAt"`  // 更新时间
}
