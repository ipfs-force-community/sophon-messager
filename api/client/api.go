package client

import (
	"github.com/filecoin-project/venus/venus-shared/api/messager"
)

var _ IMessager = (*Message)(nil)

type IMessager = messager.IMessager

type Message = messager.IMessagerStruct
