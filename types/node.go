package types

type NodeType int

const (
	_ NodeType = iota
	FullNode
	LightNode
)

type Node struct {
	ID UUID

	Name  string
	URL   string
	Token string
	Type  NodeType
}
