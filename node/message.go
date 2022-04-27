package node

import (
	"encoding/json"
	"errors"

	"github.com/luxingwen/pnet/protos"
	"github.com/luxingwen/pnet/stat"

	"github.com/luxingwen/pnet/log"

	"github.com/gogo/protobuf/proto"

	"github.com/google/uuid"
)

const (
	msgLenBytes = 4
)

func (ln *LocalNode) NewPingMessage() (*protos.Message, error) {

	msgBody := &protos.Ping{}

	buf, err := proto.Marshal(msgBody)
	if err != nil {
		return nil, err
	}

	msg := &protos.Message{
		MessageType: protos.PING,
		RoutingType: protos.DIRECT,
		MessageId:   []byte(uuid.New().String()),
		Message:     buf,
	}

	return msg, nil
}

func (ln *LocalNode) NewPingReply(replyToID []byte) (*protos.Message, error) {

	msgBody := &protos.PingReply{}

	buf, err := proto.Marshal(msgBody)
	if err != nil {
		return nil, err
	}

	msg := &protos.Message{
		MessageType: protos.PING,
		RoutingType: protos.DIRECT,
		ReplyToId:   replyToID,
		MessageId:   []byte(uuid.New().String()),
		Message:     buf,
	}
	return msg, nil
}

func (ln *LocalNode) NewStopMessage() (*protos.Message, error) {

	msgBody := &protos.Stop{}

	buf, err := proto.Marshal(msgBody)
	if err != nil {
		return nil, err
	}

	msg := &protos.Message{
		MessageType: protos.STOP,
		RoutingType: protos.DIRECT,
		MessageId:   []byte(uuid.New().String()),
		Message:     buf,
	}

	return msg, nil
}

// NewExchangeNodeMessage creates a EXCHANGE_NODE message to get node info
func (ln *LocalNode) NewExchangeNodeMessage() (*protos.Message, error) {

	msgBody := &protos.ExchangeNode{
		Node: ln.Node.Node,
	}

	buf, err := proto.Marshal(msgBody)
	if err != nil {
		return nil, err
	}

	msg := &protos.Message{
		MessageType: protos.EXCHANGE_NODE,
		RoutingType: protos.DIRECT,
		MessageId:   []byte(uuid.New().String()),
		Message:     buf,
	}

	return msg, nil
}

func (ln *LocalNode) NewExchangeNodeReply(replyToID []byte) (*protos.Message, error) {

	msgBody := &protos.ExchangeNodeReply{
		Node: ln.Node.Node,
	}

	buf, err := proto.Marshal(msgBody)
	if err != nil {
		return nil, err
	}

	msg := &protos.Message{
		MessageType: protos.EXCHANGE_NODE,
		RoutingType: protos.DIRECT,
		ReplyToId:   replyToID,
		MessageId:   []byte(uuid.New().String()),
		Message:     buf,
	}

	return msg, nil
}

func (ln *LocalNode) NewGetNeighborsMessage(destId string) (msg *protos.Message) {
	msg = &protos.Message{
		MessageType: protos.GET_NEIGHBORS,
		RoutingType: protos.RELAY,
		MessageId:   []byte(uuid.New().String()),
		DestId:      destId,
		SrcId:       ln.GetId(),
	}

	return
}

func (ln *LocalNode) NewNeighborsMessage(replyToID []byte, destid string) (msg *protos.Message, err error) {
	ns, err := ln.GetNeighbors(nil)
	if err != nil {
		return
	}
	nodes := make([]*protos.Node, 0)
	for _, item := range ns {
		nodes = append(nodes, item.Node.Node)
	}

	msgBody := &protos.Neighbors{
		Nodes: nodes,
	}

	buf, err := proto.Marshal(msgBody)
	if err != nil {
		return
	}
	msg = &protos.Message{
		MessageType: protos.BYTES,
		RoutingType: protos.RELAY,
		ReplyToId:   replyToID,
		MessageId:   []byte(uuid.New().String()),
		Message:     buf,
		SrcId:       ln.GetId(),
		DestId:      destid,
	}
	return

}

func (ln *LocalNode) NewConnnetNodeMessage(destid string, remoteNode *protos.Node) (msg *protos.Message, err error) {
	msgBody := &protos.ConnectNode{
		Node: remoteNode,
	}

	buf, err := proto.Marshal(msgBody)
	if err != nil {
		return
	}
	msg = &protos.Message{
		MessageType: protos.CONNECT_NODE,
		RoutingType: protos.RELAY,
		MessageId:   []byte(uuid.New().String()),
		Message:     buf,
		SrcId:       ln.GetId(),
		DestId:      destid,
	}
	return
}

func (ln *LocalNode) NewConnnetNodeReplayMessage(replyToID []byte, content []byte, destid string) (msg *protos.Message, err error) {
	msg = &protos.Message{
		MessageType: protos.BYTES,
		RoutingType: protos.RELAY,
		MessageId:   []byte(uuid.New().String()),
		Message:     content,
		ReplyToId:   replyToID,
		SrcId:       ln.GetId(),
		DestId:      destid,
	}
	return
}

func (ln *LocalNode) NewStopConnnetNodeMessage(destid string, remoteNode *protos.Node) (msg *protos.Message, err error) {
	msgBody := &protos.StopRemoteNode{
		Node: remoteNode,
	}

	buf, err := proto.Marshal(msgBody)
	if err != nil {
		return
	}
	msg = &protos.Message{
		MessageType: protos.STOP_REMOTENODE,
		RoutingType: protos.RELAY,
		MessageId:   []byte(uuid.New().String()),
		Message:     buf,
		SrcId:       ln.GetId(),
		DestId:      destid,
	}
	return
}

func (ln *LocalNode) NewRelayReplayMessage(replyToID []byte, content []byte, destid string) (msg *protos.Message, err error) {
	msg = &protos.Message{
		MessageType: protos.BYTES,
		RoutingType: protos.RELAY,
		MessageId:   []byte(uuid.New().String()),
		Message:     content,
		ReplyToId:   replyToID,
		SrcId:       ln.GetId(),
		DestId:      destid,
	}
	return
}

func (ln *LocalNode) NewNodeStatMessage(destid string) (msg *protos.Message) {
	msg = &protos.Message{
		MessageType: protos.NODE_STAT,
		RoutingType: protos.RELAY,
		MessageId:   []byte(uuid.New().String()),
		SrcId:       ln.GetId(),
		DestId:      destid,
	}
	return
}

func (ln *LocalNode) NewNodeStatReplayMessage(replyToID []byte, destid string) (msg *protos.Message, err error) {

	nodestat := stat.GetNodeStat()
	nodestat.PeerId = ln.GetId()
	b, err := json.Marshal(nodestat)
	if err != nil {
		return
	}

	msg = &protos.Message{
		MessageType: protos.BYTES,
		RoutingType: protos.RELAY,
		MessageId:   []byte(uuid.New().String()),
		SrcId:       ln.GetId(),
		DestId:      destid,
		ReplyToId:   replyToID,
		Message:     b,
	}
	return
}

func (ln *LocalNode) NewBraodcastMessage(data []byte) (msg *protos.Message) {
	b := &protos.Bytes{
		Data: data,
	}
	buf, err := proto.Marshal(b)
	if err != nil {
		return
	}
	msg = &protos.Message{
		MessageType: protos.BYTES,
		RoutingType: protos.BROADCAST,
		MessageId:   []byte(uuid.New().String()),
		Message:     buf,
		SrcId:       ln.GetId(),
	}
	return
}

type RemoteMessage struct {
	RemoteNode *RemoteNode
	Msg        *protos.Message
}

func NewRemoteMessage(rn *RemoteNode, msg *protos.Message) (*RemoteMessage, error) {
	remoteMsg := &RemoteMessage{
		RemoteNode: rn,
		Msg:        msg,
	}
	return remoteMsg, nil
}

func (ln *LocalNode) handleRemoteMessage(remoteMsg *RemoteMessage) error {
	if remoteMsg.RemoteNode == nil && remoteMsg.Msg.MessageType != protos.BYTES {
		return errors.New("Message is sent by local node")
	}

	switch remoteMsg.Msg.MessageType {
	case protos.PING:
		replyMsg, err := ln.NewPingReply(remoteMsg.Msg.MessageId)
		if err != nil {
			return err
		}

		err = remoteMsg.RemoteNode.SendMessageAsync(replyMsg)
		if err != nil {
			return err
		}

	case protos.EXCHANGE_NODE:
		msgBody := &protos.ExchangeNode{}
		err := proto.Unmarshal(remoteMsg.Msg.Message, msgBody)
		if err != nil {
			return err
		}

		err = remoteMsg.RemoteNode.setNode(msgBody.Node)
		if err != nil {
			remoteMsg.RemoteNode.Stop(err)
			return err
		}

		replyMsg, err := ln.NewExchangeNodeReply(remoteMsg.Msg.MessageId)
		if err != nil {
			return err
		}

		err = remoteMsg.RemoteNode.SendMessageAsync(replyMsg)
		if err != nil {
			return err
		}

	case protos.STOP:
		log.Infof("Received stop message from remote node %v", remoteMsg.RemoteNode)
		remoteMsg.RemoteNode.Stop(nil)

	case protos.BYTES:
		msgBody := &protos.Bytes{}
		err := proto.Unmarshal(remoteMsg.Msg.Message, msgBody)
		if err != nil {
			return err
		}

		data := msgBody.Data
		var shouldCallNextMiddleware bool
		for _, mw := range ln.middlewareStore.bytesReceived {
			if remoteMsg.RemoteNode == nil && remoteMsg.Msg.SrcId == ln.GetId() {
				remoteMsg.RemoteNode = &RemoteNode{Node: ln.Node}
			}
			data, shouldCallNextMiddleware = mw.Func(data, remoteMsg.Msg.MessageId, remoteMsg.Msg.SrcId, remoteMsg.Msg.Path, remoteMsg.RemoteNode)
			if !shouldCallNextMiddleware {
				break
			}
		}

	case protos.GET_NEIGHBORS:
		msg, err := ln.NewNeighborsMessage(remoteMsg.Msg.MessageId, remoteMsg.Msg.SrcId)
		if err != nil {
			return err
		}
		err = remoteMsg.RemoteNode.SendMessageAsync(msg)
		if err != nil {
			return err
		}

	case protos.CONNECT_NODE:
		cnode := &protos.ConnectNode{}
		err := proto.Unmarshal(remoteMsg.Msg.Message, cnode)
		if err != nil {
			return err
		}

		rmsg := "OK"
		_, _, err = ln.Connect(cnode.Node)
		if err != nil {
			rmsg = "Err:" + err.Error()

		}

		msg, err := ln.NewConnnetNodeReplayMessage(remoteMsg.Msg.MessageId, []byte(rmsg), remoteMsg.Msg.SrcId)
		if err != nil {
			return err
		}
		err = remoteMsg.RemoteNode.SendMessageAsync(msg)
		if err != nil {
			return err
		}

	case protos.STOP_REMOTENODE:
		cnode := &protos.StopRemoteNode{}
		err := proto.Unmarshal(remoteMsg.Msg.Message, cnode)
		if err != nil {
			return err
		}

		rm, err := ln.GetNeighbors(nil)
		if err != nil {
			return err
		}
		v, ok := rm[cnode.Node.Id]
		if ok {
			v.Stop(errors.New("STOP_REMOTENODE"))
		}

		msg, err := ln.NewRelayReplayMessage(remoteMsg.Msg.MessageId, []byte("OK"), remoteMsg.Msg.SrcId)
		if err != nil {
			return err
		}
		err = remoteMsg.RemoteNode.SendMessageAsync(msg)
		if err != nil {
			return err
		}

	case protos.NODE_STAT:
		msg, err := ln.NewNodeStatReplayMessage(remoteMsg.Msg.MessageId, remoteMsg.Msg.SrcId)
		if err != nil {
			return err
		}

		err = remoteMsg.RemoteNode.SendMessageAsync(msg)
		if err != nil {
			return err
		}

	default:
		return errors.New("Unknown message type")
	}

	return nil
}
