package lnet

import (
	"github.com/luxingwen/pnet/node"
	"github.com/luxingwen/pnet/overlay/routing"
)

const (
	RelayRoutingNumWorkers = 1
)

type BroadCastRouting struct {
	*routing.Routing
	lnet *LNet
}

func NewBroadcastRouting(localMsgChan chan<- *node.RemoteMessage, rxMsgChan <-chan *node.RemoteMessage, lnet *LNet) (*BroadCastRouting, error) {
	r, err := routing.NewRouting(localMsgChan, rxMsgChan, lnet.LocalNode)
	if err != nil {
		return nil, err
	}

	dr := &BroadCastRouting{
		Routing: r,
		lnet:    lnet,
	}

	return dr, nil
}

func (rlx *BroadCastRouting) Start() error {
	return rlx.Routing.Start(rlx, RelayRoutingNumWorkers)
}

func (rlx *BroadCastRouting) GetNodeToRoute(remoteMsg *node.RemoteMessage) (*node.LocalNode, []*node.RemoteNode, error) {

	ok, err := rlx.lnet.GetLocalNode().AddToRxCache(remoteMsg.Msg.MessageId)
	if err != nil || !ok {
		return nil, nil, err
	}

	rnNodes, err := rlx.lnet.LocalNode.GetNeighbors(func(rn *node.RemoteNode) bool {
		if remoteMsg.RemoteNode != nil && remoteMsg.RemoteNode.Id == rn.GetId() {
			return false
		}
		return true
	})

	if err != nil {
		return nil, nil, err
	}
	r := make([]*node.RemoteNode, 0)

	for _, item := range rnNodes {
		r = append(r, item)
	}

	return rlx.lnet.LocalNode, r, nil
}
