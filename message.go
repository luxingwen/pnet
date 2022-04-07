package pnet

import (
	"time"

	"github.com/luxingwen/pnet/node"
	"github.com/luxingwen/pnet/protos"
	"github.com/luxingwen/pnet/utils"

	"github.com/gogo/protobuf/proto"
)

func (pn *PNet) NewDirectBytesMessage(data []byte) (*protos.Message, error) {

	msgBody := &protos.Bytes{
		Data: data,
	}

	buf, err := proto.Marshal(msgBody)
	if err != nil {
		return nil, err
	}

	msg := &protos.Message{
		MessageType: protos.BYTES,
		RoutingType: protos.DIRECT,
		MessageId:   utils.GenId(),
		Message:     buf,
	}

	return msg, nil
}

func (pn *PNet) NewRelayBytesMessage(data []byte, srcID, key string) (*protos.Message, error) {

	msgBody := &protos.Bytes{
		Data: data,
	}

	buf, err := proto.Marshal(msgBody)
	if err != nil {
		return nil, err
	}

	msg := &protos.Message{
		MessageType: protos.BYTES,
		RoutingType: protos.RELAY,
		MessageId:   utils.GenId(),
		Message:     buf,
		SrcId:       srcID,
		DestId:      key,
	}

	return msg, nil
}

func (pn *PNet) SendBytesDirectAsync(data []byte, remoteNode *node.RemoteNode) error {
	msg, err := pn.NewDirectBytesMessage(data)
	if err != nil {
		return err
	}

	return remoteNode.SendMessageAsync(msg)
}

func (pn *PNet) SendBytesDirectSync(data []byte, remoteNode *node.RemoteNode) ([]byte, *node.RemoteNode, error) {
	return pn.SendBytesDirectSyncWithTimeout(data, remoteNode, 0)
}

func (pn *PNet) SendBytesDirectSyncWithTimeout(data []byte, remoteNode *node.RemoteNode, replyTimeout time.Duration) ([]byte, *node.RemoteNode, error) {
	msg, err := pn.NewDirectBytesMessage(data)
	if err != nil {
		return nil, nil, err
	}

	reply, err := remoteNode.SendMessageSync(msg, replyTimeout)
	if err != nil {
		return nil, nil, err
	}

	replyBody := &protos.Bytes{}
	err = proto.Unmarshal(reply.Msg.Message, replyBody)
	if err != nil {
		return nil, reply.RemoteNode, err
	}

	return replyBody.Data, reply.RemoteNode, nil
}

func (pn *PNet) SendBytesDirectReply(replyToID, data []byte, remoteNode *node.RemoteNode) error {
	msg, err := pn.NewDirectBytesMessage(data)
	if err != nil {
		return err
	}

	msg.ReplyToId = replyToID

	return remoteNode.SendMessageAsync(msg)
}

func (pn *PNet) SendBytesRelayAsync(data []byte, key string) (bool, error) {
	msg, err := pn.NewRelayBytesMessage(data, pn.GetLocalNode().Id, key)
	if err != nil {
		return false, err
	}

	return pn.SendMessageAsync(msg, protos.RELAY)
}

func (pn *PNet) SendBytesRelaySync(data []byte, key string) ([]byte, string, error) {
	return pn.SendBytesRelaySyncWithTimeout(data, key, 0)
}

func (pn *PNet) SendBytesRelaySyncWithTimeout(data []byte, key string, replyTimeout time.Duration) ([]byte, string, error) {
	msg, err := pn.NewRelayBytesMessage(data, pn.GetLocalNode().Id, key)
	if err != nil {
		return nil, "", err
	}

	reply, _, err := pn.SendMessageSync(msg, protos.RELAY, replyTimeout)
	if err != nil {
		return nil, "", err
	}

	replyBody := &protos.Bytes{}
	err = proto.Unmarshal(reply.Message, replyBody)
	if err != nil {
		return nil, reply.SrcId, err
	}

	return replyBody.Data, reply.SrcId, nil
}

func (pn *PNet) SendBytesRelayReply(replyToID, data []byte, key string) (bool, error) {
	msg, err := pn.NewRelayBytesMessage(data, pn.GetLocalNode().Id, key)
	if err != nil {
		return false, err
	}

	msg.ReplyToId = replyToID

	return pn.SendMessageAsync(msg, protos.RELAY)
}
