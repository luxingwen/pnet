package lnet

import (
	"fmt"
	"pnet/log"
	"pnet/node"
)

func (ln *LNet) handleRemoteMessage(remoteMsg *node.RemoteMessage) (bool, error) {
	if remoteMsg.RemoteNode == nil {
		return true, nil
	}

	for _, routingType := range ln.LocalNode.LocalReciveMsgCacheRoutingType {
		if routingType == remoteMsg.Msg.RoutingType {

			added, err := ln.LocalNode.AddToRxCache(remoteMsg.Msg.MessageId)
			if err != nil {
				log.Errorf("Add msg id %x to rx cache error: %v", remoteMsg.Msg.MessageId, err)
				return false, nil
			}
			if !added {
				fmt.Println("added: ", added, "msgid : ", string(remoteMsg.Msg.MessageId), "routingType:", routingType)
				return false, nil
			}
		}
	}

	return true, nil
}
