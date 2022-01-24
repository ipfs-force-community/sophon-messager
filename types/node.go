package types

import (
	"github.com/filecoin-project/venus/venus-shared/types/messager"
)

type NodeType = messager.NodeType

const (
	FullNode  = messager.FullNode
	LightNode = messager.LightNode
)

type Node = messager.Node
