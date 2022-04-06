package node

import (
	"errors"
	"pnet/protos"

	"pnet/log"

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
			data, shouldCallNextMiddleware = mw.Func(data, remoteMsg.Msg.MessageId, remoteMsg.Msg.SrcId, remoteMsg.Msg.Path, remoteMsg.RemoteNode)
			if !shouldCallNextMiddleware {
				break
			}
		}

	default:
		return errors.New("Unknown message type")
	}

	return nil
}
