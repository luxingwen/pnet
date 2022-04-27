package lnet

import (
	"github.com/luxingwen/pnet/log"
	"github.com/luxingwen/pnet/overlay"
	"github.com/luxingwen/pnet/overlay/routing"
	"github.com/luxingwen/pnet/protos"

	"github.com/luxingwen/pnet/node"
)

const (
	numWorkers = 1

	joinRetries = 3
)

type LNet struct {
	*overlay.Overlay
}

func NewLNet(localNode *node.LocalNode) (*LNet, error) {
	ovl, err := overlay.NewOverlay(localNode)
	if err != nil {
		return nil, err
	}

	ln := &LNet{
		Overlay: ovl,
	}

	directRxMsgChan, err := localNode.GetRxMsgChan(protos.DIRECT)
	if err != nil {
		return nil, err
	}
	directRouting, err := routing.NewDirectRouting(ovl.LocalMsgChan, directRxMsgChan, localNode)
	if err != nil {
		return nil, err
	}
	err = ovl.AddRouter(protos.DIRECT, directRouting)
	if err != nil {
		return nil, err
	}

	relayRxMsgChan, err := localNode.GetRxMsgChan(protos.RELAY)
	if err != nil {
		return nil, err
	}
	relayRouting, err := NewRelayRouting(ovl.LocalMsgChan, relayRxMsgChan, ln)
	if err != nil {
		return nil, err
	}
	err = ovl.AddRouter(protos.RELAY, relayRouting)
	if err != nil {
		return nil, err
	}

	braodcastRxMsgChan, err := localNode.GetRxMsgChan(protos.BROADCAST)
	if err != nil {
		return nil, err
	}

	broadcastRouting, err := NewBroadcastRouting(ovl.LocalMsgChan, braodcastRxMsgChan, ln)

	if err != nil {
		return nil, err
	}

	err = ovl.AddRouter(protos.BROADCAST, broadcastRouting)
	if err != nil {
		return nil, err
	}

	return ln, nil
}

func (ln *LNet) Stop(err error) {
	ln.StopOnce.Do(func() {

		ln.LocalNode.Stop(err)

		ln.LifeCycle.Stop()

		ln.StopRouters(err)

	})
}

func (ln *LNet) Start() (err error) {
	ln.StartOnce.Do(func() {

		err = ln.StartRouters()
		if err != nil {
			ln.Stop(err)
			return
		}

		for i := 0; i < numWorkers; i++ {
			go ln.handleMsg()
		}

		err = ln.LocalNode.Start()
		if err != nil {
			ln.Stop(err)
			return
		}

	})
	return
}

func (ln *LNet) Join(seedNodeAddr string) (remoteNode *node.RemoteNode, err error) {
	return ln.Connect(&protos.Node{Addr: seedNodeAddr})
}

func (ln *LNet) Connect(n *protos.Node) (remoteNode *node.RemoteNode, err error) {

	remoteNode, _, err = ln.LocalNode.Connect(n)
	if err != nil {
		return
	}

	return
}

func (ln *LNet) handleMsg() {
	var remoteMsg *node.RemoteMessage
	var shouldLocalNodeHandleMsg bool
	var err error

	for {
		if ln.IsStopped() {
			return
		}

		remoteMsg = <-ln.LocalMsgChan

		shouldLocalNodeHandleMsg, err = ln.handleRemoteMessage(remoteMsg)
		if err != nil {
			log.Errorf("LNET handle remote message error: %v", err)
			continue
		}

		if shouldLocalNodeHandleMsg {
			err = ln.LocalNode.HandleRemoteMessage(remoteMsg)
			if err != nil {
				log.Errorf("Local node handle remote message error: %v", err)
				continue
			}
		}
	}
}
