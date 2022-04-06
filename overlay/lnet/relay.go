package lnet

import (
	"pnet/node"
	"pnet/overlay/routing"
	"strings"
)

const (
	RelayRoutingNumWorkers = 1
)

type RelayRouting struct {
	*routing.Routing
	lnet *LNet
}

func NewRelayRouting(localMsgChan chan<- *node.RemoteMessage, rxMsgChan <-chan *node.RemoteMessage, lnet *LNet) (*RelayRouting, error) {
	r, err := routing.NewRouting(localMsgChan, rxMsgChan, lnet.LocalNode)
	if err != nil {
		return nil, err
	}

	dr := &RelayRouting{
		Routing: r,
		lnet:    lnet,
	}

	return dr, nil
}

func (rlx *RelayRouting) Start() error {
	return rlx.Routing.Start(rlx, RelayRoutingNumWorkers)
}

func (rlx *RelayRouting) GetNodeToRoute(remoteMsg *node.RemoteMessage) (*node.LocalNode, []*node.RemoteNode, error) {

	if remoteMsg.Msg.DestId == rlx.lnet.LocalNode.Id {
		return rlx.lnet.LocalNode, nil, nil
	}

	if remoteMsg.Msg.Path != "" {
		rs1, err := rlx.getNodeByPath(remoteMsg)
		if err == nil && len(rs1) > 0 {
			return nil, rs1, nil
		}
	}

	mBroadCast := make(map[string]string, 0)

	for _, item := range remoteMsg.Msg.BroadcastNodes {
		mBroadCast[item] = item
	}
	rnNodes, err := rlx.lnet.LocalNode.GetNeighbors(func(rn *node.RemoteNode) bool {
		if remoteMsg.Msg.DestId == rlx.lnet.LocalNode.Id || rn.Type != "client" {
			if _, ok := mBroadCast[rn.Id]; !ok {
				return true
			}
			return false
		}
		return false
	})
	if err != nil {
		return nil, nil, err
	}

	rns := make([]*node.RemoteNode, 0)

	if v, ok := rnNodes[string(remoteMsg.Msg.DestId)]; ok {
		rns = append(rns, v)
		return nil, rns, nil
	}

	for _, item := range rnNodes {
		rns = append(rns, item)
	}

	return nil, rns, nil
}

func (rlx *RelayRouting) getNodeByPath(remoteMsg *node.RemoteMessage) (rns []*node.RemoteNode, err error) {

	list := strings.Split(remoteMsg.Msg.Path, "/")

	idx := 0
	for i, item := range list {
		if item == rlx.lnet.LocalNode.Id {
			idx = i + 1
			break
		}
	}

	if idx >= len(list) {
		return
	}

	foundId := list[idx]

	rnNodes, err := rlx.lnet.LocalNode.GetNeighbors(func(rn *node.RemoteNode) bool {
		if foundId == rn.Id {
			return true
		}
		return false
	})

	if err != nil {
		return
	}

	for _, item := range rnNodes {
		rns = append(rns, item)
	}

	return

}
