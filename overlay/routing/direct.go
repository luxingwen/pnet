package routing

import (
	"errors"

	"github.com/luxingwen/pnet/node"
)

const (
	DirectRoutingNumWorkers = 1
)

// 直接链接的路由
type DirectRouting struct {
	*Routing
}

func NewDirectRouting(localMsgChan chan<- *node.RemoteMessage, rxMsgChan <-chan *node.RemoteMessage, localNode *node.LocalNode) (*DirectRouting, error) {
	r, err := NewRouting(localMsgChan, rxMsgChan, localNode)
	if err != nil {
		return nil, err
	}

	dr := &DirectRouting{
		Routing: r,
	}

	return dr, nil
}

func (dr *DirectRouting) Start() error {
	return dr.Routing.Start(dr, DirectRoutingNumWorkers)
}

func (dr *DirectRouting) GetNodeToRoute(remoteMsg *node.RemoteMessage) (*node.LocalNode, []*node.RemoteNode, error) {
	if remoteMsg.RemoteNode == nil {
		return nil, nil, errors.New("Message is sent by local node")
	}

	return remoteMsg.RemoteNode.LocalNode, nil, nil
}
