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

	p1 := newPnet("p1", hostname, 50001)
	p2 := newPnet("p2", hostname, 50002)
	p3 := newPnet("p3", hostname, 50003)
	p4 := newPnet("p4", hostname, 50004)
	p5 := newPnet("p5", hostname, 50005)
	p6 := newPnet("p6", hostname, 50006)

	p2.Join(p1.GetLocalNode().Addr)
	p3.Join(p2.GetLocalNode().Addr)
	p4.Join(p3.GetLocalNode().Addr)
	p5.Join(p4.GetLocalNode().Addr)
	p6.Join(p5.GetLocalNode().Addr)

	time.Sleep(time.Second * 3)

	reply, _, err := p2.SendMessageSync(p2.GetLocalNode().NewNodeStatMessage(p6.GetLocalNode().GetId()), protos.RELAY, 0)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	fmt.Println(string(reply.Message))

}
