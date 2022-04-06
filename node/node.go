package node

import (
	"fmt"
	"pnet/common"
	"pnet/protos"
	"sync"
)

type Node struct {
	sync.RWMutex
	*protos.Node
	common.LifeCycle
}

func newNode(n *protos.Node) *Node {
	node := &Node{
		Node: n,
	}
	return node
}

func NewNode(id string, name string, addr string) *Node {
	n := &protos.Node{
		Id:   id,
		Name: name,
		Addr: addr,
	}
	return newNode(n)
}

func (n *Node) String() string {
	return fmt.Sprintf("%s@%s", n.Id, n.Addr)
}

func (n *Node) SetType(typ string) {
	n.Type = typ
}
