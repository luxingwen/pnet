package node

import (
	"fmt"
	"sync"

	"github.com/luxingwen/pnet/common"
	"github.com/luxingwen/pnet/protos"
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
