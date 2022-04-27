package routing

import (
	"errors"
	"strings"
	"time"

	"github.com/luxingwen/pnet/common"
	"github.com/luxingwen/pnet/log"
	"github.com/luxingwen/pnet/node"
	"github.com/luxingwen/pnet/utils"
)

type Router interface {
	Start() error
	Stop(error)
	ApplyMiddleware(interface{}) error
	GetNodeToRoute(remoteMsg *node.RemoteMessage) (localNode *node.LocalNode, remoteNodes []*node.RemoteNode, err error)
	SendMessage(router Router, remoteMsg *node.RemoteMessage, hasReply bool, replyTimeout time.Duration) (replyChan <-chan *node.RemoteMessage, success bool, err error)
}

type Routing struct {
	localMsgChan chan<- *node.RemoteMessage
	rxMsgChan    <-chan *node.RemoteMessage
	*middlewareStore
	common.LifeCycle
	*node.LocalNode
}

func NewRouting(localMsgChan chan<- *node.RemoteMessage, rxMsgChan <-chan *node.RemoteMessage, localNode *node.LocalNode) (*Routing, error) {
	r := &Routing{
		localMsgChan:    localMsgChan,
		rxMsgChan:       rxMsgChan,
		middlewareStore: newMiddlewareStore(),
		LocalNode:       localNode,
	}
	return r, nil
}

func (r *Routing) Start(router Router, numWorkers int) error {
	r.StartOnce.Do(func() {
		for i := 0; i < numWorkers; i++ {
			go r.handleMsg(router)
		}
	})

	return nil
}

func (r *Routing) Stop(err error) {
	r.StopOnce.Do(func() {
		if err != nil {
			log.Warningf("Routing stops because of error: %s", err)
		} else {
			log.Infof("Routing stops")
		}

		r.LifeCycle.Stop()
	})
}

func (r *Routing) SendMessage(router Router, remoteMsg *node.RemoteMessage, hasReply bool, replyTimeout time.Duration) (<-chan *node.RemoteMessage, bool, error) {
	var shouldCallNextMiddleware bool
	success := false

	localNode, remoteNodes, err := router.GetNodeToRoute(remoteMsg)
	if err != nil {
		return nil, false, err
	}

	for _, mw := range r.middlewareStore.remoteMessageRouted {
		remoteMsg, localNode, remoteNodes, shouldCallNextMiddleware = mw.Func(remoteMsg, localNode, remoteNodes)
		if remoteMsg == nil || !shouldCallNextMiddleware {
			break
		}
	}

	if remoteMsg == nil {
		return nil, false, nil
	}

	if localNode == nil && len(remoteNodes) == 0 {
		//fmt.Println("loclid:", r.LocalNode.GetId(), " body:", string(remoteMsg.Msg.Message), " pth->", remoteMsg.Msg.Path, " srcid:", remoteMsg.Msg.SrcId, "dest id: ", remoteMsg.Msg.DestId, " type:", remoteMsg.Msg.GetRoutingType().String())
		return nil, false, errors.New("No node to route")
	}

	if localNode != nil {
		err = r.sendMessageToLocalNode(remoteMsg, localNode)
		if err != nil {
			return nil, false, err
		}
		success = true
	}

	//	fmt.Println("remotenodes:", remoteNodes)

	var replyChan <-chan *node.RemoteMessage
	errs := utils.NewErrors()

	if !strings.Contains(remoteMsg.Msg.Path, r.LocalNode.Id) {
		if remoteMsg.Msg.Path != "" {
			remoteMsg.Msg.Path += "/"
		}
		remoteMsg.Msg.Path += r.LocalNode.Id
	}

	mbrocast := make(map[string]string, 0)

	for _, item := range remoteMsg.Msg.BroadcastNodes {
		mbrocast[item] = item
	}

	if _, ok := mbrocast[r.LocalNode.Id]; !ok {
		remoteMsg.Msg.BroadcastNodes = append(remoteMsg.Msg.BroadcastNodes, r.LocalNode.Id)
	}

	for _, remoteNode := range remoteNodes {
		if _, ok := mbrocast[r.LocalNode.Id]; !ok {
			remoteMsg.Msg.BroadcastNodes = append(remoteMsg.Msg.BroadcastNodes, remoteNode.Id)
		}
	}

	for _, remoteNode := range remoteNodes {

		if hasReply && replyChan == nil {
			replyChan, err = remoteNode.SendMessage(remoteMsg.Msg, true, replyTimeout)
		} else {
			_, err = remoteNode.SendMessage(remoteMsg.Msg, false, 0)
		}

		if err != nil {
			errs = append(errs, err)
		} else {
			success = true
		}
	}

	if !success {
		return nil, false, errs.Merged()
	}

	return replyChan, success, nil
}

func (r *Routing) sendMessageToLocalNode(remoteMsg *node.RemoteMessage, localNode *node.LocalNode) error {
	var shouldCallNextMiddleware bool

	for _, mw := range r.middlewareStore.remoteMessageReceived {
		remoteMsg, shouldCallNextMiddleware = mw.Func(remoteMsg)
		if remoteMsg == nil || !shouldCallNextMiddleware {
			break
		}
	}

	if remoteMsg == nil {
		return nil
	}

	for _, routingType := range r.LocalNode.LocalReciveMsgCacheRoutingType {
		if routingType == remoteMsg.Msg.RoutingType {
			added, err := r.LocalNode.AddToRxCache(remoteMsg.Msg.MessageId)
			if err != nil {
				log.Errorf("Add msg id %x to rx cache error: %v", remoteMsg.Msg.MessageId, err)
				return nil
			}
			if !added {
				return nil
			}
		}
	}

	if len(remoteMsg.Msg.ReplyToId) > 0 {

		replyChan, ok := localNode.GetReplyChan(remoteMsg.Msg.ReplyToId)
		if ok && replyChan != nil {
			select {
			case replyChan <- remoteMsg:
			default:
				log.Warning("Reply chan unavailable or full, discarding msg, node id:", r.LocalNode.GetId(), " len(replyChan): ", len(replyChan), " cap(replyChan):", cap(replyChan), " message type:", remoteMsg.Msg.MessageType.String(), "msg:", string(remoteMsg.Msg.Message), "replay id:", string(remoteMsg.Msg.ReplyToId))
			}
		}
		return nil
	}

	select {
	case r.localMsgChan <- remoteMsg:
	default:
		log.Warning("Router local msg chan full, discarding msg")
	}

	return nil
}

func (r *Routing) handleMsg(router Router) {
	var remoteMsg *node.RemoteMessage
	var shouldCallNextMiddleware bool
	var err error

	for {
		if r.IsStopped() {
			return
		}

		remoteMsg = <-r.rxMsgChan

		for _, mw := range r.middlewareStore.remoteMessageArrived {
			remoteMsg, shouldCallNextMiddleware = mw.Func(remoteMsg)
			if remoteMsg == nil || !shouldCallNextMiddleware {
				break
			}
		}

		if remoteMsg == nil {
			continue
		}

		_, _, err = r.SendMessage(router, remoteMsg, false, 0)
		if err != nil {
			log.Warningf("%s Route message error: %v, srcid:%s, descid:%s,  path:%s, broadcasts:%v", r.LocalNode.GetId(), err, remoteMsg.Msg.SrcId, remoteMsg.Msg.DestId,
				remoteMsg.Msg.Path, remoteMsg.Msg.BroadcastNodes)
		}
	}
}
