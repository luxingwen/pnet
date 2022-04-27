package main

import (
	"fmt"
	"time"

	"github.com/luxingwen/pnet"

	"github.com/luxingwen/pnet/config"
	"github.com/luxingwen/pnet/log"
	"github.com/luxingwen/pnet/node"
	"github.com/luxingwen/pnet/protos"
)

func newPnet(id string, name string, port uint16) *pnet.PNet {

	cfg := config.DefaultConfig()
	cfg.Hostname = "127.0.0.1"
	cfg.Port = port
	pn, err := pnet.NewPNet(id, cfg)
	if err != nil {
		panic(err)
	}

	pn.ApplyMiddleware(node.LocalNodeStarted{func(lc *node.LocalNode) bool {
		lc.SetReady(true)
		return true
	}, 0})

	pn.ApplyMiddleware(node.BytesReceived{func(msg, msgID []byte, srcID, rpath string, remoteNode *node.RemoteNode) ([]byte, bool) {
		log.Infof("Receive message \"%s\" from %s by %s , path: %s ", string(msg), srcID, remoteNode.Id, rpath)
		pn.SendBytesRelayReply(msgID, []byte("receive send res:"+rpath), srcID)
		return nil, true
	}, 0})

	err = pn.Start()
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(time.Second)
		if pn.GetLocalNode().IsReady() {
			return pn
		}
	}

	return pn
}

func main() {
	hostname := "127.0.0.1"

	p1 := newPnet("p1", hostname, 50011)

	reply, _, err := p1.SendMessageSync(p1.GetLocalNode().NewNodeStatMessage(p1.GetLocalNode().GetId()), protos.RELAY, 0)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	fmt.Println(string(reply.Message))

}
