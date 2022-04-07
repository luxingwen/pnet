package node

import (
	"errors"

	"github.com/luxingwen/pnet/protos"

	"github.com/luxingwen/pnet/middleware"
)

type BytesReceived struct {
	Func     func(data, msgID []byte, srcID, rpath string, remoteNode *RemoteNode) ([]byte, bool)
	Priority int32
}

type LocalNodeWillStart struct {
	Func     func(*LocalNode) bool
	Priority int32
}

type LocalNodeStarted struct {
	Func     func(*LocalNode) bool
	Priority int32
}

type LocalNodeWillStop struct {
	Func     func(*LocalNode) bool
	Priority int32
}

type LocalNodeStopped struct {
	Func     func(*LocalNode) bool
	Priority int32
}

type WillConnectToNode struct {
	Func     func(*protos.Node) (bool, bool)
	Priority int32
}

type RemoteNodeConnected struct {
	Func     func(*RemoteNode) bool
	Priority int32
}

type RemoteNodeReady struct {
	Func     func(*RemoteNode) bool
	Priority int32
}

type RemoteNodeDisconnected struct {
	Func     func(*RemoteNode) bool
	Priority int32
}

type MessageEncoded struct {
	Func     func(*RemoteNode, []byte) ([]byte, bool)
	Priority int32
}

type MessageWillDecode struct {
	Func     func(*RemoteNode, []byte) ([]byte, bool)
	Priority int32
}

type middlewareStore struct {
	bytesReceived          []BytesReceived
	localNodeWillStart     []LocalNodeWillStart
	localNodeStarted       []LocalNodeStarted
	localNodeWillStop      []LocalNodeWillStop
	localNodeStopped       []LocalNodeStopped
	willConnectToNode      []WillConnectToNode
	remoteNodeConnected    []RemoteNodeConnected
	remoteNodeReady        []RemoteNodeReady
	remoteNodeDisconnected []RemoteNodeDisconnected
	messageEncoded         []MessageEncoded
	messageWillDecode      []MessageWillDecode
}

func newMiddlewareStore() *middlewareStore {
	return &middlewareStore{
		bytesReceived:          make([]BytesReceived, 0),
		localNodeWillStart:     make([]LocalNodeWillStart, 0),
		localNodeStarted:       make([]LocalNodeStarted, 0),
		localNodeWillStop:      make([]LocalNodeWillStop, 0),
		localNodeStopped:       make([]LocalNodeStopped, 0),
		willConnectToNode:      make([]WillConnectToNode, 0),
		remoteNodeConnected:    make([]RemoteNodeConnected, 0),
		remoteNodeReady:        make([]RemoteNodeReady, 0),
		remoteNodeDisconnected: make([]RemoteNodeDisconnected, 0),
		messageEncoded:         make([]MessageEncoded, 0),
		messageWillDecode:      make([]MessageWillDecode, 0),
	}
}

func (store *middlewareStore) ApplyMiddleware(mw interface{}) error {
	switch mw := mw.(type) {
	case BytesReceived:
		if mw.Func == nil {
			return errors.New("middleware function is nil")
		}
		store.bytesReceived = append(store.bytesReceived, mw)
		middleware.Sort(store.bytesReceived)
	case LocalNodeWillStart:
		if mw.Func == nil {
			return errors.New("middleware function is nil")
		}
		store.localNodeWillStart = append(store.localNodeWillStart, mw)
		middleware.Sort(store.localNodeWillStart)
	case LocalNodeStarted:
		if mw.Func == nil {
			return errors.New("middleware function is nil")
		}
		store.localNodeStarted = append(store.localNodeStarted, mw)
		middleware.Sort(store.localNodeStarted)
	case LocalNodeWillStop:
		if mw.Func == nil {
			return errors.New("middleware function is nil")
		}
		store.localNodeWillStop = append(store.localNodeWillStop, mw)
		middleware.Sort(store.localNodeWillStop)
	case LocalNodeStopped:
		if mw.Func == nil {
			return errors.New("middleware function is nil")
		}
		store.localNodeStopped = append(store.localNodeStopped, mw)
		middleware.Sort(store.localNodeStopped)
	case WillConnectToNode:
		if mw.Func == nil {
			return errors.New("middleware function is nil")
		}
		store.willConnectToNode = append(store.willConnectToNode, mw)
		middleware.Sort(store.willConnectToNode)
	case RemoteNodeConnected:
		if mw.Func == nil {
			return errors.New("middleware function is nil")
		}
		store.remoteNodeConnected = append(store.remoteNodeConnected, mw)
		middleware.Sort(store.remoteNodeConnected)
	case RemoteNodeReady:
		if mw.Func == nil {
			return errors.New("middleware function is nil")
		}
		store.remoteNodeReady = append(store.remoteNodeReady, mw)
		middleware.Sort(store.remoteNodeReady)
	case RemoteNodeDisconnected:
		if mw.Func == nil {
			return errors.New("middleware function is nil")
		}
		store.remoteNodeDisconnected = append(store.remoteNodeDisconnected, mw)
		middleware.Sort(store.remoteNodeDisconnected)
	case MessageEncoded:
		if mw.Func == nil {
			return errors.New("middleware function is nil")
		}
		store.messageEncoded = append(store.messageEncoded, mw)
		middleware.Sort(store.messageEncoded)
	case MessageWillDecode:
		if mw.Func == nil {
			return errors.New("middleware function is nil")
		}
		store.messageWillDecode = append(store.messageWillDecode, mw)
		middleware.Sort(store.messageWillDecode)
	default:
		return errors.New("unknown middleware type")
	}

	return nil
}
