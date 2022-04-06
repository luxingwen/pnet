package lnet

import "pnet/node"

func (ln *LNet) handleRemoteMessage(remoteMsg *node.RemoteMessage) (bool, error) {
	if remoteMsg.RemoteNode == nil {
		return true, nil
	}

	return true, nil
}
