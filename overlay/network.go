package overlay

import (
	"time"

	"github.com/luxingwen/pnet/node"
	"github.com/luxingwen/pnet/overlay/routing"
	"github.com/luxingwen/pnet/protos"
)

type Network interface {
	Start() error
	Stop(error)
	Join(seedNodeAddr string) (*node.RemoteNode, error)
	GetLocalNode() *node.LocalNode
	GetRouters() []routing.Router
	SendMessageAsync(msg *protos.Message, routingType protos.RoutingType) (success bool, err error)
	SendMessageSync(msg *protos.Message, routingType protos.RoutingType, replyTimeout time.Duration) (reply *protos.Message, success bool, err error)
}
