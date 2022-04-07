package overlay

import (
	"errors"
	"fmt"
	"time"

	"github.com/luxingwen/pnet/common"
	"github.com/luxingwen/pnet/node"
	"github.com/luxingwen/pnet/overlay/routing"
	"github.com/luxingwen/pnet/protos"
)

type Overlay struct {
	LocalNode    *node.LocalNode
	LocalMsgChan chan *node.RemoteMessage
	routers      map[protos.RoutingType]routing.Router
	common.LifeCycle
}

func NewOverlay(localNode *node.LocalNode) (*Overlay, error) {
	if localNode == nil {
		return nil, errors.New("Local node is nil")
	}

	overlay := &Overlay{
		LocalNode:    localNode,
		LocalMsgChan: make(chan *node.RemoteMessage, localNode.OverlayLocalMsgChanLen),
		routers:      make(map[protos.RoutingType]routing.Router),
	}
	return overlay, nil
}

func (ovl *Overlay) GetLocalNode() *node.LocalNode {
	return ovl.LocalNode
}

func (ovl *Overlay) AddRouter(routingType protos.RoutingType, router routing.Router) error {
	_, ok := ovl.routers[routingType]
	if ok {
		return fmt.Errorf("Router for type %v is already added", routingType)
	}
	ovl.routers[routingType] = router
	return nil
}

func (ovl *Overlay) GetRouter(routingType protos.RoutingType) (routing.Router, error) {
	router, ok := ovl.routers[routingType]
	if !ok {
		return nil, fmt.Errorf("Router for type %v has not been added yet", routingType)
	}
	return router, nil
}

func (ovl *Overlay) GetRouters() []routing.Router {
	routers := make([]routing.Router, 0)
	for _, router := range ovl.routers {
		routers = append(routers, router)
	}
	return routers
}

func (ovl *Overlay) SetRouter(routingType protos.RoutingType, router routing.Router) {
	ovl.routers[routingType] = router
}

func (ovl *Overlay) StartRouters() error {
	for _, router := range ovl.routers {
		err := router.Start()
		if err != nil {
			return nil
		}
	}

	return nil
}

func (ovl *Overlay) StopRouters(err error) {
	for _, router := range ovl.routers {
		router.Stop(err)
	}
}

func (ovl *Overlay) SendMessage(msg *protos.Message, routingType protos.RoutingType, hasReply bool, replyTimeout time.Duration) (<-chan *node.RemoteMessage, bool, error) {
	router, err := ovl.GetRouter(routingType)
	if err != nil {
		return nil, false, err
	}

	return router.SendMessage(router, &node.RemoteMessage{Msg: msg}, hasReply, replyTimeout)
}

func (ovl *Overlay) SendMessageAsync(msg *protos.Message, routingType protos.RoutingType) (bool, error) {
	_, success, err := ovl.SendMessage(msg, routingType, false, 0)
	return success, err
}

func (ovl *Overlay) SendMessageSync(msg *protos.Message, routingType protos.RoutingType, replyTimeout time.Duration) (*protos.Message, bool, error) {
	if replyTimeout == 0 {
		replyTimeout = ovl.LocalNode.DefaultReplyTimeout
	}

	replyChan, success, err := ovl.SendMessage(msg, routingType, true, replyTimeout)
	if !success {
		return nil, false, err
	}

	select {
	case replyMsg := <-replyChan:
		return replyMsg.Msg, true, nil
	case <-time.After(replyTimeout):
		return nil, true, errors.New("Wait for reply timeout")
	}
}
