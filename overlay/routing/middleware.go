package routing

import (
	"errors"
	"pnet/middleware"
	"pnet/node"
)

type RemoteMessageArrived struct {
	Func     func(*node.RemoteMessage) (*node.RemoteMessage, bool)
	Priority int32
}

type RemoteMessageRouted struct {
	Func     func(*node.RemoteMessage, *node.LocalNode, []*node.RemoteNode) (*node.RemoteMessage, *node.LocalNode, []*node.RemoteNode, bool)
	Priority int32
}

type RemoteMessageReceived struct {
	Func     func(*node.RemoteMessage) (*node.RemoteMessage, bool)
	Priority int32
}

type middlewareStore struct {
	remoteMessageArrived  []RemoteMessageArrived
	remoteMessageRouted   []RemoteMessageRouted
	remoteMessageReceived []RemoteMessageReceived
}

func newMiddlewareStore() *middlewareStore {
	return &middlewareStore{
		remoteMessageArrived:  make([]RemoteMessageArrived, 0),
		remoteMessageRouted:   make([]RemoteMessageRouted, 0),
		remoteMessageReceived: make([]RemoteMessageReceived, 0),
	}
}

func (store *middlewareStore) ApplyMiddleware(mw interface{}) error {
	switch mw := mw.(type) {
	case RemoteMessageArrived:
		if mw.Func == nil {
			return errors.New("middleware function is nil")
		}
		store.remoteMessageArrived = append(store.remoteMessageArrived, mw)
		middleware.Sort(store.remoteMessageArrived)
	case RemoteMessageRouted:
		if mw.Func == nil {
			return errors.New("middleware function is nil")
		}
		store.remoteMessageRouted = append(store.remoteMessageRouted, mw)
		middleware.Sort(store.remoteMessageRouted)
	case RemoteMessageReceived:
		if mw.Func == nil {
			return errors.New("middleware function is nil")
		}
		store.remoteMessageReceived = append(store.remoteMessageReceived, mw)
		middleware.Sort(store.remoteMessageReceived)
	default:
		return errors.New("unknown middleware type")
	}

	return nil
}
