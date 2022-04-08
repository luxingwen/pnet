package lnet

import (
	"github.com/luxingwen/pnet/node"
)

func (ln *LNet) handleRemoteMessage(remoteMsg *node.RemoteMessage) (bool, error) {
	if remoteMsg.RemoteNode == nil {
		return true, nil
	}

	return true, nil
}
